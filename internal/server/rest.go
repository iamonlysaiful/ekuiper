// Copyright 2021 EMQ Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lf-edge/ekuiper/internal/conf"
	"github.com/lf-edge/ekuiper/internal/meta"
	"github.com/lf-edge/ekuiper/internal/plugin/native"
	"github.com/lf-edge/ekuiper/internal/service"
	"github.com/lf-edge/ekuiper/pkg/api"
	"github.com/lf-edge/ekuiper/pkg/ast"
	"github.com/lf-edge/ekuiper/pkg/errorx"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	ContentType     = "Content-Type"
	ContentTypeJSON = "application/json"
)

type statementDescriptor struct {
	Sql string `json:"sql,omitempty"`
}

func decodeStatementDescriptor(reader io.ReadCloser) (statementDescriptor, error) {
	sd := statementDescriptor{}
	err := json.NewDecoder(reader).Decode(&sd)
	// Problems decoding
	if err != nil {
		return sd, fmt.Errorf("Error decoding the statement descriptor: %v", err)
	}
	return sd, nil
}

// Handle applies the specified error and error concept tot he HTTP response writer
func handleError(w http.ResponseWriter, err error, prefix string, logger api.Logger) {
	message := prefix
	if message != "" {
		message += ": "
	}
	message += err.Error()
	logger.Error(message)
	var ec int
	switch e := err.(type) {
	case *errorx.Error:
		switch e.Code() {
		case errorx.NOT_FOUND:
			ec = http.StatusNotFound
		default:
			ec = http.StatusBadRequest
		}
	default:
		ec = http.StatusBadRequest
	}
	http.Error(w, message, ec)
}

func jsonResponse(i interface{}, w http.ResponseWriter, logger api.Logger) {
	w.Header().Add(ContentType, ContentTypeJSON)
	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		handleError(w, err, "", logger)
	}
}

func createRestServer(ip string, port int) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)
	r.HandleFunc("/streams", streamsHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/streams/{name}", streamHandler).Methods(http.MethodGet, http.MethodDelete, http.MethodPut)
	r.HandleFunc("/tables", tablesHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/tables/{name}", tableHandler).Methods(http.MethodGet, http.MethodDelete, http.MethodPut)
	r.HandleFunc("/rules", rulesHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/rules/{name}", ruleHandler).Methods(http.MethodDelete, http.MethodGet, http.MethodPut)
	r.HandleFunc("/rules/{name}/status", getStatusRuleHandler).Methods(http.MethodGet)
	r.HandleFunc("/rules/{name}/start", startRuleHandler).Methods(http.MethodPost)
	r.HandleFunc("/rules/{name}/stop", stopRuleHandler).Methods(http.MethodPost)
	r.HandleFunc("/rules/{name}/restart", restartRuleHandler).Methods(http.MethodPost)
	r.HandleFunc("/rules/{name}/topo", getTopoRuleHandler).Methods(http.MethodGet)

	r.HandleFunc("/plugins/sources", sourcesHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/plugins/sources/prebuild", prebuildSourcePlugins).Methods(http.MethodGet)
	r.HandleFunc("/plugins/sources/{name}", sourceHandler).Methods(http.MethodDelete, http.MethodGet)

	r.HandleFunc("/plugins/sinks", sinksHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/plugins/sinks/prebuild", prebuildSinkPlugins).Methods(http.MethodGet)
	r.HandleFunc("/plugins/sinks/{name}", sinkHandler).Methods(http.MethodDelete, http.MethodGet)
	r.HandleFunc("/plugins/functions", functionsHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/plugins/functions/prebuild", prebuildFuncsPlugins).Methods(http.MethodGet)
	r.HandleFunc("/plugins/functions/{name}", functionHandler).Methods(http.MethodDelete, http.MethodGet)
	r.HandleFunc("/plugins/functions/{name}/register", functionRegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/plugins/udfs", functionsListHandler).Methods(http.MethodGet)
	r.HandleFunc("/plugins/udfs/{name}", functionsGetHandler).Methods(http.MethodGet)

	r.HandleFunc("/metadata/functions", functionsMetaHandler).Methods(http.MethodGet)

	r.HandleFunc("/metadata/sinks", sinksMetaHandler).Methods(http.MethodGet)
	r.HandleFunc("/metadata/sinks/{name}", newSinkMetaHandler).Methods(http.MethodGet)

	r.HandleFunc("/metadata/sources", sourcesMetaHandler).Methods(http.MethodGet)
	r.HandleFunc("/metadata/sources/yaml/{name}", sourceConfHandler).Methods(http.MethodGet)
	r.HandleFunc("/metadata/sources/{name}", sourceMetaHandler).Methods(http.MethodGet)
	r.HandleFunc("/metadata/sources/{name}/confKeys", sourceConfKeysHandler).Methods(http.MethodGet)
	r.HandleFunc("/metadata/sources/{name}/confKeys/{confKey}", sourceConfKeyHandler).Methods(http.MethodDelete, http.MethodPost)
	r.HandleFunc("/metadata/sources/{name}/confKeys/{confKey}/field", sourceConfKeyFieldsHandler).Methods(http.MethodDelete, http.MethodPost)

	r.HandleFunc("/services", servicesHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/services/functions", serviceFunctionsHandler).Methods(http.MethodGet)
	r.HandleFunc("/services/functions/{name}", serviceFunctionHandler).Methods(http.MethodGet)
	r.HandleFunc("/services/{name}", serviceHandler).Methods(http.MethodDelete, http.MethodGet, http.MethodPut)

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", ip, port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 60 * 5,
		ReadTimeout:  time.Second * 60 * 5,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CORS(handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"}))(r),
	}
	server.SetKeepAlivesEnabled(false)
	return server
}

type information struct {
	Version       string `json:"version"`
	Os            string `json:"os"`
	UpTimeSeconds int64  `json:"upTimeSeconds"`
}

//The handler for root
func rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	switch r.Method {
	case http.MethodGet, http.MethodPost:
		w.WriteHeader(http.StatusOK)
		info := new(information)
		info.Version = version
		info.UpTimeSeconds = time.Now().Unix() - startTimeStamp
		info.Os = runtime.GOOS
		byteInfo, _ := json.Marshal(info)
		w.Write(byteInfo)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func sourcesManageHandler(w http.ResponseWriter, r *http.Request, st ast.StreamType) {
	defer r.Body.Close()
	switch r.Method {
	case http.MethodGet:
		content, err := streamProcessor.ShowStream(st)
		if err != nil {
			handleError(w, err, fmt.Sprintf("%s command error", strings.Title(ast.StreamTypeMap[st])), logger)
			return
		}
		jsonResponse(content, w, logger)
	case http.MethodPost:
		v, err := decodeStatementDescriptor(r.Body)
		if err != nil {
			handleError(w, err, "Invalid body", logger)
			return
		}
		content, err := streamProcessor.ExecStreamSql(v.Sql)
		if err != nil {
			handleError(w, err, fmt.Sprintf("%s command error", strings.Title(ast.StreamTypeMap[st])), logger)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(content))
	}
}

func sourceManageHandler(w http.ResponseWriter, r *http.Request, st ast.StreamType) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	switch r.Method {
	case http.MethodGet:
		content, err := streamProcessor.DescStream(name, st)
		if err != nil {
			handleError(w, err, fmt.Sprintf("describe %s error", ast.StreamTypeMap[st]), logger)
			return
		}
		jsonResponse(content, w, logger)
	case http.MethodDelete:
		content, err := streamProcessor.DropStream(name, st)
		if err != nil {
			handleError(w, err, fmt.Sprintf("delete %s error", ast.StreamTypeMap[st]), logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	case http.MethodPut:
		v, err := decodeStatementDescriptor(r.Body)
		if err != nil {
			handleError(w, err, "Invalid body", logger)
			return
		}
		content, err := streamProcessor.ExecReplaceStream(v.Sql, st)
		if err != nil {
			handleError(w, err, fmt.Sprintf("%s command error", strings.Title(ast.StreamTypeMap[st])), logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}
}

//list or create streams
func streamsHandler(w http.ResponseWriter, r *http.Request) {
	sourcesManageHandler(w, r, ast.TypeStream)
}

//describe or delete a stream
func streamHandler(w http.ResponseWriter, r *http.Request) {
	sourceManageHandler(w, r, ast.TypeStream)
}

//list or create tables
func tablesHandler(w http.ResponseWriter, r *http.Request) {
	sourcesManageHandler(w, r, ast.TypeTable)
}

func tableHandler(w http.ResponseWriter, r *http.Request) {
	sourceManageHandler(w, r, ast.TypeTable)
}

//list or create rules
func rulesHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	switch r.Method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleError(w, err, "Invalid body", logger)
			return
		}
		r, err := ruleProcessor.ExecCreate("", string(body))
		var result string
		if err != nil {
			handleError(w, err, "Create rule error", logger)
			return
		} else {
			result = fmt.Sprintf("Rule %s was created successfully.", r.Id)
		}
		//Start the rule
		rs, err := createRuleState(r)
		if err != nil {
			result = err.Error()
		} else {
			err = doStartRule(rs)
			if err != nil {
				result = err.Error()
			}
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(result))
	case http.MethodGet:
		content, err := getAllRulesWithStatus()
		if err != nil {
			handleError(w, err, "Show rules error", logger)
			return
		}
		jsonResponse(content, w, logger)
	}
}

//describe or delete a rule
func ruleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	switch r.Method {
	case http.MethodGet:
		rule, err := ruleProcessor.GetRuleByName(name)
		if err != nil {
			handleError(w, err, "describe rule error", logger)
			return
		}
		jsonResponse(rule, w, logger)
	case http.MethodDelete:
		deleteRule(name)
		content, err := ruleProcessor.ExecDrop(name)
		if err != nil {
			handleError(w, err, "delete rule error", logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	case http.MethodPut:
		_, err := ruleProcessor.GetRuleByName(name)
		if err != nil {
			handleError(w, err, "not found this rule", logger)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleError(w, err, "Invalid body", logger)
			return
		}

		r, err := ruleProcessor.ExecUpdate(name, string(body))
		var result string
		if err != nil {
			handleError(w, err, "Update rule error", logger)
			return
		} else {
			result = fmt.Sprintf("Rule %s was updated successfully.", r.Id)
		}

		err = restartRule(name)
		if err != nil {
			handleError(w, err, "restart rule error", logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}
}

//get status of a rule
func getStatusRuleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	content, err := getRuleStatus(name)
	if err != nil {
		handleError(w, err, "get rule status error", logger)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

//start a rule
func startRuleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	err := startRule(name)
	if err != nil {
		handleError(w, err, "start rule error", logger)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Rule %s was started", name)))
}

//stop a rule
func stopRuleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	result := stopRule(name)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

//restart a rule
func restartRuleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	err := restartRule(name)
	if err != nil {
		handleError(w, err, "restart rule error", logger)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Rule %s was restarted", name)))
}

//get topo of a rule
func getTopoRuleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]

	content, err := getRuleTopo(name)
	if err != nil {
		handleError(w, err, "get rule topo error", logger)
		return
	}
	w.Header().Set(ContentType, ContentTypeJSON)
	w.Write([]byte(content))
}

func pluginsHandler(w http.ResponseWriter, r *http.Request, t native.PluginType) {
	pluginManager := native.GetManager()
	defer r.Body.Close()
	switch r.Method {
	case http.MethodGet:
		content := pluginManager.List(t)
		jsonResponse(content, w, logger)
	case http.MethodPost:
		sd := native.NewPluginByType(t)
		err := json.NewDecoder(r.Body).Decode(sd)
		// Problems decoding
		if err != nil {
			handleError(w, err, fmt.Sprintf("Invalid body: Error decoding the %s plugin json", native.PluginTypes[t]), logger)
			return
		}
		err = pluginManager.Register(t, sd)
		if err != nil {
			handleError(w, err, fmt.Sprintf("%s plugins create command error", native.PluginTypes[t]), logger)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("%s plugin %s is created", native.PluginTypes[t], sd.GetName())))
	}
}

func pluginHandler(w http.ResponseWriter, r *http.Request, t native.PluginType) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]
	cb := r.URL.Query().Get("stop")
	pluginManager := native.GetManager()
	switch r.Method {
	case http.MethodDelete:
		r := cb == "1"
		err := pluginManager.Delete(t, name, r)
		if err != nil {
			handleError(w, err, fmt.Sprintf("delete %s plugin %s error", native.PluginTypes[t], name), logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		result := fmt.Sprintf("%s plugin %s is deleted", native.PluginTypes[t], name)
		if r {
			result = fmt.Sprintf("%s and Kuiper will be stopped", result)
		} else {
			result = fmt.Sprintf("%s and Kuiper must restart for the change to take effect.", result)
		}
		w.Write([]byte(result))
	case http.MethodGet:
		j, ok := pluginManager.GetPluginInfo(t, name)
		if !ok {
			handleError(w, errorx.NewWithCode(errorx.NOT_FOUND, "not found"), fmt.Sprintf("describe %s plugin %s error", native.PluginTypes[t], name), logger)
			return
		}
		jsonResponse(j, w, logger)
	}
}

//list or create source plugin
func sourcesHandler(w http.ResponseWriter, r *http.Request) {
	pluginsHandler(w, r, native.SOURCE)
}

//delete a source plugin
func sourceHandler(w http.ResponseWriter, r *http.Request) {
	pluginHandler(w, r, native.SOURCE)
}

//list or create sink plugin
func sinksHandler(w http.ResponseWriter, r *http.Request) {
	pluginsHandler(w, r, native.SINK)
}

//delete a sink plugin
func sinkHandler(w http.ResponseWriter, r *http.Request) {
	pluginHandler(w, r, native.SINK)
}

//list or create function plugin
func functionsHandler(w http.ResponseWriter, r *http.Request) {
	pluginsHandler(w, r, native.FUNCTION)
}

//list all user defined functions in all function plugins
func functionsListHandler(w http.ResponseWriter, r *http.Request) {
	pluginManager := native.GetManager()
	content := pluginManager.ListSymbols()
	jsonResponse(content, w, logger)
}

func functionsGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	pluginManager := native.GetManager()
	j, ok := pluginManager.GetPluginBySymbol(native.FUNCTION, name)
	if !ok {
		handleError(w, errorx.NewWithCode(errorx.NOT_FOUND, "not found"), fmt.Sprintf("describe function %s error", name), logger)
		return
	}
	jsonResponse(map[string]string{"name": name, "plugin": j}, w, logger)
}

//delete a function plugin
func functionHandler(w http.ResponseWriter, r *http.Request) {
	pluginHandler(w, r, native.FUNCTION)
}

type functionList struct {
	Functions []string `json:"functions,omitempty"`
}

// register function list for function plugin. If a plugin exports multiple functions, the function list must be registered
// either by create or register. If the function plugin has been loaded because of auto load through so file, the function
// list MUST be registered by this API or only the function with the same name as the plugin can be used.
func functionRegisterHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]
	pluginManager := native.GetManager()
	_, ok := pluginManager.GetPluginInfo(native.FUNCTION, name)
	if !ok {
		handleError(w, errorx.NewWithCode(errorx.NOT_FOUND, "not found"), fmt.Sprintf("register %s plugin %s error", native.PluginTypes[native.FUNCTION], name), logger)
		return
	}
	sd := functionList{}
	err := json.NewDecoder(r.Body).Decode(&sd)
	// Problems decoding
	if err != nil {
		handleError(w, err, fmt.Sprintf("Invalid body: Error decoding the function list json %s", r.Body), logger)
		return
	}
	err = pluginManager.RegisterFuncs(name, sd.Functions)
	if err != nil {
		handleError(w, err, fmt.Sprintf("function plugins %s regiser functions error", name), logger)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("function plugin %s function list is registered", name)))
}

func prebuildSourcePlugins(w http.ResponseWriter, r *http.Request) {
	prebuildPluginsHandler(w, r, native.SOURCE)
}

func prebuildSinkPlugins(w http.ResponseWriter, r *http.Request) {
	prebuildPluginsHandler(w, r, native.SINK)
}

func prebuildFuncsPlugins(w http.ResponseWriter, r *http.Request) {
	prebuildPluginsHandler(w, r, native.FUNCTION)
}

func isOffcialDockerImage() bool {
	if strings.ToLower(os.Getenv("MAINTAINER")) != "emqx.io" {
		return false
	}
	return true
}

func prebuildPluginsHandler(w http.ResponseWriter, r *http.Request, t native.PluginType) {
	emsg := "It's strongly recommended to install plugins at official released Debian Docker images. If you choose to proceed to install plugin, please make sure the plugin is already validated in your own build."
	if !isOffcialDockerImage() {
		handleError(w, fmt.Errorf(emsg), "", logger)
		return
	} else if runtime.GOOS == "linux" {
		osrelease, err := Read()
		if err != nil {
			logger.Infof("")
			return
		}
		prettyName := strings.ToUpper(osrelease["PRETTY_NAME"])
		os := "debian"
		if strings.Contains(prettyName, "DEBIAN") {
			hosts := conf.Config.Basic.PluginHosts
			ptype := "sources"
			if t == native.SINK {
				ptype = "sinks"
			} else if t == native.FUNCTION {
				ptype = "functions"
			}
			if err, plugins := fetchPluginList(hosts, ptype, os, runtime.GOARCH); err != nil {
				handleError(w, err, "", logger)
			} else {
				jsonResponse(plugins, w, logger)
			}
		} else {
			handleError(w, fmt.Errorf(emsg), "", logger)
			return
		}
	} else {
		handleError(w, fmt.Errorf(emsg), "", logger)
	}
}

func fetchPluginList(hosts, ptype, os, arch string) (err error, result map[string]string) {
	if hosts == "" || ptype == "" || os == "" {
		logger.Errorf("Invalid parameter value: hosts %s, ptype %s or os: %s should not be empty.", hosts, ptype, os)
		return fmt.Errorf("Invalid configruation for plugin host in kuiper.yaml."), nil
	}
	result = make(map[string]string)
	hostsArr := strings.Split(hosts, ",")
	for _, host := range hostsArr {
		host := strings.Trim(host, " ")
		tmp := []string{host, "kuiper-plugins", version, os, ptype}
		//The url is similar to http://host:port/kuiper-plugins/0.9.1/debian/sinks/
		url := strings.Join(tmp, "/")
		timeout := time.Duration(30 * time.Second)
		client := &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err := client.Get(url)
		logger.Infof("Trying to fetch plugins from url: %s\n", url)

		if err != nil {
			return err, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Cannot fetch plugin list from %s, with status error: %v", url, resp.StatusCode), nil
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err, nil
		}
		plugins := extractFromHtml(string(data), arch)
		for _, p := range plugins {
			//If already existed, using the existed.
			if _, ok := result[p]; !ok {
				result[p] = url + "/" + p + "_" + arch + ".zip"
			}
			logger.Debugf("Plugin %s, download address is %s\n", p, result[p])
		}
	}
	return
}

func extractFromHtml(content, arch string) []string {
	plugins := []string{}
	htmlTokens := html.NewTokenizer(strings.NewReader(content))
loop:
	for {
		tt := htmlTokens.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.StartTagToken:
			t := htmlTokens.Token()
			isAnchor := t.Data == "a"
			if isAnchor {
				found := false
				for _, prop := range t.Attr {
					if strings.ToUpper(prop.Key) == "HREF" {
						if strings.HasSuffix(prop.Val, "_"+arch+".zip") {
							if index := strings.LastIndex(prop.Val, "_"); index != -1 {
								plugins = append(plugins, prop.Val[0:index])
							}
						}
						found = true
					}
				}
				if !found {
					logger.Infof("Invalid plugin download link %s", t)
				}
			}
		}
	}
	return plugins
}

//list sink plugin
func sinksMetaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sinks := meta.GetSinks()
	jsonResponse(sinks, w, logger)
	return
}

//Get sink metadata when creating rules
func newSinkMetaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	pluginName := vars["name"]

	language := getLanguage(r)
	ptrMetadata, err := meta.GetSinkMeta(pluginName, language)
	if err != nil {
		handleError(w, err, "", logger)
		return
	}
	jsonResponse(ptrMetadata, w, logger)
}

//list functions
func functionsMetaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sinks := meta.GetFunctions()
	jsonResponse(sinks, w, logger)
	return
}

//list source plugin
func sourcesMetaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ret := meta.GetSources()
	if nil != ret {
		jsonResponse(ret, w, logger)
		return
	}
}

//Get source metadata when creating stream
func sourceMetaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	pluginName := vars["name"]
	language := getLanguage(r)
	ret, err := meta.GetSourceMeta(pluginName, language)
	if err != nil {
		handleError(w, err, "", logger)
		return
	}
	if nil != ret {
		jsonResponse(ret, w, logger)
		return
	}
}

//Get source yaml
func sourceConfHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	pluginName := vars["name"]
	language := getLanguage(r)
	ret, err := meta.GetSourceConf(pluginName, language)
	if err != nil {
		handleError(w, err, "", logger)
		return
	} else {
		w.Write(ret)
	}
}

//Get confKeys of the source metadata
func sourceConfKeysHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	pluginName := vars["name"]
	ret := meta.GetSourceConfKeys(pluginName)
	if nil != ret {
		jsonResponse(ret, w, logger)
		return
	}
}

//Add  del confkey
func sourceConfKeyHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	var ret interface{}
	var err error
	vars := mux.Vars(r)
	pluginName := vars["name"]
	confKey := vars["confKey"]
	language := getLanguage(r)
	switch r.Method {
	case http.MethodDelete:
		err = meta.DelSourceConfKey(pluginName, confKey, language)
	case http.MethodPost:
		v, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleError(w, err, "Invalid body", logger)
			return
		}
		err = meta.AddSourceConfKey(pluginName, confKey, language, v)
	}
	if err != nil {
		handleError(w, err, "", logger)
		return
	}
	if nil != ret {
		jsonResponse(ret, w, logger)
		return
	}
}

//Del and Update field of confkey
func sourceConfKeyFieldsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var ret interface{}
	var err error
	vars := mux.Vars(r)
	pluginName := vars["name"]
	confKey := vars["confKey"]
	v, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError(w, err, "Invalid body", logger)
		return
	}

	language := getLanguage(r)
	switch r.Method {
	case http.MethodDelete:
		err = meta.DelSourceConfKeyField(pluginName, confKey, language, v)
	case http.MethodPost:
		err = meta.AddSourceConfKeyField(pluginName, confKey, language, v)
	}
	if err != nil {
		handleError(w, err, "", logger)
		return
	}
	if nil != ret {
		jsonResponse(ret, w, logger)
		return
	}
}
func getLanguage(r *http.Request) string {
	language := r.Header.Get("Content-Language")
	if 0 == len(language) {
		language = "en_US"
	}
	return language
}

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	serviceManager := service.GetManager()
	switch r.Method {
	case http.MethodGet:
		content, err := serviceManager.List()
		if err != nil {
			handleError(w, err, "service list command error", logger)
			return
		}
		jsonResponse(content, w, logger)
	case http.MethodPost:
		sd := &service.ServiceCreationRequest{}
		err := json.NewDecoder(r.Body).Decode(sd)
		// Problems decoding
		if err != nil {
			handleError(w, err, "Invalid body: Error decoding the %s service request payload", logger)
			return
		}
		err = serviceManager.Create(sd)
		if err != nil {
			handleError(w, err, "service create command error", logger)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("service %s is created", sd.Name)))
	}
}

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	name := vars["name"]
	serviceManager := service.GetManager()
	switch r.Method {
	case http.MethodDelete:
		err := serviceManager.Delete(name)
		if err != nil {
			handleError(w, err, fmt.Sprintf("delete service %s error", name), logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		result := fmt.Sprintf("service %s is deleted", name)
		w.Write([]byte(result))
	case http.MethodGet:
		j, err := serviceManager.Get(name)
		if err != nil {
			handleError(w, errorx.NewWithCode(errorx.NOT_FOUND, "not found"), fmt.Sprintf("describe service %s error", name), logger)
			return
		}
		jsonResponse(j, w, logger)
	case http.MethodPut:
		sd := &service.ServiceCreationRequest{}
		err := json.NewDecoder(r.Body).Decode(sd)
		// Problems decoding
		if err != nil {
			handleError(w, err, "Invalid body: Error decoding the %s service request payload", logger)
			return
		}
		sd.Name = name
		err = serviceManager.Update(sd)
		if err != nil {
			handleError(w, err, "service update command error", logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("service %s is updated", sd.Name)))
	}
}

func serviceFunctionsHandler(w http.ResponseWriter, r *http.Request) {
	serviceManager := service.GetManager()
	content, err := serviceManager.ListFunctions()
	if err != nil {
		handleError(w, err, "service list command error", logger)
		return
	}
	jsonResponse(content, w, logger)
}

func serviceFunctionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	serviceManager := service.GetManager()
	j, err := serviceManager.GetFunction(name)
	if err != nil {
		handleError(w, errorx.NewWithCode(errorx.NOT_FOUND, "not found"), fmt.Sprintf("describe function %s error", name), logger)
		return
	}
	jsonResponse(j, w, logger)
}

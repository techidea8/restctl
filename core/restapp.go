package core

import (
	"embed"
	"errors"
	"fmt"
	"linker/utils"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	log "github.com/techidea8/restctl/pkg/log"
)

type RestApp struct {
	router           *mux.Router
	Host             string
	Port             int
	Config           *AppCfg
	methodMaplocker  *sync.RWMutex
	moduleMaplocker  *sync.RWMutex
	processHandleMap map[string]reflect.Value
	ctrlMap          map[string]reflect.Value
	restCtrl         *RestCtrl
}

func NewRestApp(appCfg *AppCfg) *RestApp {
	return &RestApp{
		router:           mux.NewRouter(),
		Host:             appCfg.Host,
		Port:             appCfg.Port,
		Config:           appCfg,
		methodMaplocker:  &sync.RWMutex{},
		moduleMaplocker:  &sync.RWMutex{},
		processHandleMap: map[string]reflect.Value{},
		ctrlMap:          map[string]reflect.Value{},
		restCtrl:         &RestCtrl{},
	}
}

func (r *RestApp) HandlerFunc(patern string, fun http.HandlerFunc, methods ...string) *RestApp {
	r.router.HandleFunc(patern, fun).Methods(methods...)
	return r
}

func (r *RestApp) Post(patern string, fun http.HandlerFunc) *RestApp {
	return r.HandlerFunc(patern, fun, "POST")
}

func (r *RestApp) Get(patern string, fun http.HandlerFunc) *RestApp {
	return r.HandlerFunc(patern, fun, "GET")
}
func (r *RestApp) Put(patern string, fun http.HandlerFunc) *RestApp {
	return r.HandlerFunc(patern, fun, "PUT")
}
func (r *RestApp) Options(patern string, fun http.HandlerFunc) *RestApp {
	return r.HandlerFunc(patern, fun, "OPTIONS")
}
func (r *RestApp) Delete(patern string, fun http.HandlerFunc) *RestApp {
	return r.HandlerFunc(patern, fun, "DELETE")
}
func (r *RestApp) Any(patern string, fun http.HandlerFunc) *RestApp {
	return r.HandlerFunc(patern, fun, "DELETE", "POST", "GET", "OPTIONS", "PUT")
}

// 使用中间件
func (r *RestApp) UseMiddleware(middlewareFuncs ...mux.MiddlewareFunc) *RestApp {
	//
	r.router.Use(middlewareFuncs...)
	return r
}

// 初始化数据库
func (r *RestApp) FsMap(fs embed.FS, patern, root string) *RestApp {
	fsHandler := FSHandler{Fs: fs, Root: root, Patern: patern}
	r.router.PathPrefix(patern).Handler(fsHandler)
	log.Debugf("map file %s to %s", root, patern)
	return r
}
func Register(apps ...interface{}) {
	for _, v := range apps {
		mapIRestCtrl[v] = true
	}
}
func (app *RestApp) Register(apps ...interface{}) *RestApp {
	Register(apps...)
	return app
}

// 这个不能放到App 里面去
var mapIRestCtrl map[interface{}]bool = map[interface{}]bool{}

// 这个用来处理图片验证密码
func (app *RestApp) findmethod(module, action string) (result reflect.Value, err error) {
	defer utils.TimeCost("findmethod cost ->" + module + "/" + action)()
	// /:module/:action
	// 首字母小写
	methodName := utils.UcFirst(action)
	key := module + "/" + action
	app.methodMaplocker.Lock()
	method, ok := app.processHandleMap[key]
	app.methodMaplocker.Unlock()
	if !ok {
		ctrlName := utils.UcFirst(module)
		app.moduleMaplocker.Lock()
		ctrl, ok := app.ctrlMap[ctrlName]
		app.moduleMaplocker.Unlock()
		if !ok {
			err = errors.New("该服务不存在")
			return
		}
		method = ctrl.MethodByName(methodName)
		if !method.IsValid() {
			err = errors.New("请求的操作:" + methodName + "不存在")
			return
		}
		app.methodMaplocker.Lock()
		app.processHandleMap[key] = method
		app.methodMaplocker.Unlock()
	}
	result = method
	err = nil
	return
}

// 这个用来处理图片验证密码
func (app *RestApp) Dispatcher(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	module := params["module"]
	action := params["action"]
	defer utils.TimeCost("dispatch cost ->" + module + "/" + action)()
	method, err := app.findmethod(module, action)
	if err != nil {
		app.restCtrl.RespFail(w, err.Error())
	} else {
		args := []reflect.Value{reflect.ValueOf(w), reflect.ValueOf(req)}
		log.Debugf("rest->%s/%s", module, action)
		method.Call(args)
	}
}

// 使用中间件
func (app *RestApp) Start() {
	//启动
	for ctrl := range mapIRestCtrl {
		// 获取结构体实例的反射类型对象
		dataType := reflect.TypeOf(ctrl)
		dataValue := reflect.ValueOf(ctrl)
		moduleFullName := dataType.String()
		moduleNameArr := strings.Split(moduleFullName, ".")
		moduleName := moduleNameArr[len(moduleNameArr)-1]
		app.moduleMaplocker.Lock()
		app.ctrlMap[moduleName] = dataValue
		app.moduleMaplocker.Unlock()
		log.Infof("register ctrl %s", moduleName)
	}
	app.HandlerFunc("/{module}/{action}", app.Dispatcher, "POST", "GET", "OPTIONS")
	srv := &http.Server{
		Handler:      app.router,
		Addr:         fmt.Sprintf("%s:%d", app.Host, app.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof("App Run at O(∩_∩)O %s:%d\n", app.Host, app.Port)
	log.Fatal(srv.ListenAndServe())
}

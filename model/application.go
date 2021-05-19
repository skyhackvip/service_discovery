package model

import (
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"log"
	"sync"
	"time"
)

type Application struct {
	appid           string
	instances       map[string]*Instance
	latestTimestamp int64
	lock            sync.RWMutex
}

func NewApplication(appid string) *Application {
	return &Application{
		appid:     appid,
		instances: make(map[string]*Instance),
	}
}

//注册app add instance
//返回 *Instance 实例信息
//返回 bool true 已有实例升级 false 新增实例
func (app *Application) AddInstance(in *Instance, latestTimestamp int64) (*Instance, bool) {
	app.lock.Lock()
	defer app.lock.Unlock()
	appIns, ok := app.instances[in.Hostname]
	if ok { //exist
		in.UpTimestamp = appIns.UpTimestamp
		//dirtytimestamp
		if in.DirtyTimestamp < appIns.DirtyTimestamp {
			log.Println("register exist dirty timestamp ")
			in = appIns
		}
	}
	//add or update instances
	app.instances[in.Hostname] = in
	app.upLatestTimestamp(latestTimestamp)
	returnIns := new(Instance)
	*returnIns = *in
	return returnIns, !ok
}

//续约
func (app *Application) Renew(hostname string) (*Instance, bool) {
	app.lock.Lock()
	defer app.lock.Unlock()
	appIn, ok := app.instances[hostname]
	if !ok {
		return nil, ok
	}
	//modify renew time
	appIn.RenewTimestamp = time.Now().UnixNano()
	//get copy
	return copyInstance(appIn), true
}

//取消
func (app *Application) Cancel(hostname string, latestTimestamp int64) (*Instance, bool, int) {
	newInstance := new(Instance)
	app.lock.Lock()
	defer app.lock.Unlock()
	appIn, ok := app.instances[hostname]
	if !ok {
		return nil, ok, 0
	}
	//delete hostname
	delete(app.instances, hostname)
	appIn.LatestTimestamp = latestTimestamp
	app.upLatestTimestamp(latestTimestamp)
	*newInstance = *appIn
	return newInstance, true, len(app.instances)
}

//获取所有*Instance
func (app *Application) GetAllInstances() []*Instance {
	app.lock.RLock()
	defer app.lock.RUnlock()
	rs := make([]*Instance, 0, len(app.instances))
	for _, instance := range app.instances {
		newInstance := new(Instance)
		*newInstance = *instance
		rs = append(rs, newInstance)
	}
	return rs
}

//获取*Instance信息
//status=1 return up 实例
func (app *Application) GetInstance(status uint32, latestTime int64) (*FetchData, *errcode.Error) {
	app.lock.RLock()
	defer app.lock.RUnlock()
	if latestTime >= app.latestTimestamp { //not modify
		return nil, errcode.NotModified
	}
	fetchData := FetchData{
		Instances:       make([]*Instance, 0),
		LatestTimestamp: app.latestTimestamp,
	}
	var exists bool
	for _, instance := range app.instances {
		if status&instance.Status > 0 {
			exists = true
			newInstance := copyInstance(instance)
			fetchData.Instances = append(fetchData.Instances, newInstance)
		}
	}
	if !exists {
		return nil, errcode.NotFound
	}
	return &fetchData, nil
}

func (app *Application) GetInstanceLen() int {
	app.lock.RLock()
	instanceLen := len(app.instances)
	app.lock.RUnlock()
	return instanceLen
}

//update app latest_timestamp
func (app *Application) upLatestTimestamp(latestTimestamp int64) {
	if latestTimestamp <= app.latestTimestamp { //already latest
		latestTimestamp = app.latestTimestamp + 1 //increase
	}
	app.latestTimestamp = latestTimestamp
}

//deep copy
func copyInstance(src *Instance) *Instance {
	dst := new(Instance)
	*dst = *src
	//copy addrs
	dst.Addrs = make([]string, len(src.Addrs))
	for i, addr := range src.Addrs {
		dst.Addrs[i] = addr
	}
	return dst
}

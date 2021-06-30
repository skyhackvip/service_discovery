package model

import (
	"fmt"
	"github.com/skyhackvip/service_discovery/configs"
	//"github.com/skyhackvip/service_discovery/global"
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Registry struct {
	apps map[string]*Application //key=appid+env
	lock sync.RWMutex
	gd   *Guard //protect
}

func NewRegistry() *Registry {
	r := &Registry{
		apps: make(map[string]*Application),
		gd:   new(Guard),
	}
	go r.evictTask()
	return r
}

//register
func (r *Registry) Register(instance *Instance, latestTimestamp int64) (*Application, *errcode.Error) {
	key := getKey(instance.AppId, instance.Env)
	r.lock.RLock()
	app, ok := r.apps[key]
	r.lock.RUnlock()
	if !ok { //new app
		app = NewApplication(instance.AppId)
	}

	//add instance
	_, isNew := app.AddInstance(instance, latestTimestamp)
	if isNew {
		r.gd.incrNeed()
	}
	//add into registry apps
	r.lock.Lock()
	r.apps[key] = app
	r.lock.Unlock()
	return app, nil
}

//renew
func (r *Registry) Renew(env, appid, hostname string) (*Instance, *errcode.Error) {
	//find app
	app, ok := r.getApplication(appid, env)
	if !ok {
		return nil, errcode.NotFound
	}
	//modify instance renewtime
	in, ok := app.Renew(hostname)
	if !ok {
		return nil, errcode.NotFound
	}
	r.gd.incrCount()
	return in, nil
}

//cancel
func (r *Registry) Cancel(env, appid, hostname string, latestTimestamp int64) (*Instance, *errcode.Error) {
	//find app
	app, ok := r.getApplication(appid, env)
	if !ok {
		return nil, errcode.NotFound
	}
	instance, ok, insLen := app.Cancel(hostname, latestTimestamp)
	if !ok {
		return nil, errcode.NotFound
	}
	//if instances is empty, delete app from apps
	if insLen == 0 {
		r.lock.Lock()
		delete(r.apps, getKey(appid, env))
		r.lock.Unlock()
	}
	r.gd.decrNeed()
	return instance, nil
}

//get by appname
func (r *Registry) Fetch(env, appid string, status uint32, latestTime int64) (*FetchData, *errcode.Error) {
	app, ok := r.getApplication(appid, env)
	if !ok {
		return nil, errcode.NotFound
	}
	return app.GetInstance(status, latestTime) //err = not modify
}

//get all key=appid, value=[]*Instance
func (r *Registry) FetchAll() map[string][]*Instance {
	apps := r.getAllApplications()
	rs := make(map[string][]*Instance)
	for _, app := range apps {
		rs[app.appid] = append(rs[app.appid], app.GetAllInstances()...)
	}
	return rs
}

func (r *Registry) getApplication(appid, env string) (*Application, bool) {
	key := getKey(appid, env)
	r.lock.RLock()
	app, ok := r.apps[key]
	r.lock.RUnlock()
	return app, ok
}

func (r *Registry) getAllApplications() []*Application {
	r.lock.RLock()
	defer r.lock.RUnlock()
	apps := make([]*Application, 0, len(r.apps))
	for _, app := range r.apps {
		apps = append(apps, app)
	}
	return apps
}

func (r *Registry) evictTask() {
	ticker := time.Tick(configs.CheckEvictInterval)
	resetTicker := time.Tick(configs.ResetGuardNeedCountInterval)
	for {
		select {
		case <-ticker:
			log.Println("### registry evict task every 60s ###")
			r.gd.storeLastCount()
			r.evict()
		case <-resetTicker:
			log.Println("### registry reset task every 15m ###")
			var count int64
			for _, app := range r.getAllApplications() {
				count += int64(app.GetInstanceLen())
			}
			r.gd.setNeed(count)
		}
	}
}

//evict expired instance
func (r *Registry) evict() {
	now := time.Now().UnixNano()
	var expiredInstances []*Instance
	apps := r.getAllApplications()
	var registryLen int
	protectStatus := r.gd.selfProtectStatus()
	for _, app := range apps {
		registryLen += app.GetInstanceLen()
		allInstances := app.GetAllInstances()
		for _, instance := range allInstances {
			delta := now - instance.RenewTimestamp
			if !protectStatus && delta > int64(configs.InstanceExpireDuration) ||
				delta > int64(configs.InstanceMaxExpireDuration) {
				expiredInstances = append(expiredInstances, instance)
			}
		}
	}
	evictionLimit := registryLen - int(float64(registryLen)*configs.SelfProtectThreshold)
	expiredLen := len(expiredInstances)
	if expiredLen > evictionLimit {
		expiredLen = evictionLimit
	}
	if expiredLen == 0 {
		return
	}
	for i := 0; i < expiredLen; i++ {
		j := i + rand.Intn(len(expiredInstances)-i)
		expiredInstances[i], expiredInstances[j] = expiredInstances[j], expiredInstances[i]
		expiredInstance := expiredInstances[i]
		r.Cancel(expiredInstance.Env, expiredInstance.AppId, expiredInstance.Hostname, now)
		//todo 取消广播
		//global.Discovery.Nodes.Load().(*Nodes).Replicate(configs.Cancel, expiredInstance)
		log.Printf("### evict instance (%v, %v,%v)###\n", expiredInstance.Env, expiredInstance.AppId, expiredInstance.Hostname)

	}
}

func getKey(appid, env string) string {
	return fmt.Sprintf("%s-%s", appid, env)
}

package config

import (
	"fmt"
	"reflect"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigManagerStartup(t *testing.T) {
	// setup
	epCfg := EndpointConfig{
		Name:        "Foo",
		ServiceHost: "http://localhost:8080",
		Uri:         "/foo",
		Method:      "POST",
		QueueConfig: map[interface{}]interface{}{
			"queuename": "test.foo",
		},
	}
	cfgMgr := NewWatcher(&StaticRepo{Endpoints: []EndpointConfig{epCfg}}, 5)

	// when
	err := cfgMgr.processProviders()
	found := <-cfgMgr.Updates

	// then
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(epCfg, found.Config), fmt.Sprintf("\n   %+v\n!= %+v", epCfg, found.Config))
}
func TestConfigManagerChange(t *testing.T) {
	// setup
	epCfg := EndpointConfig{
		Name:        "Foo",
		ServiceHost: "http://localhost:8080",
		Uri:         "/foo",
		Method:      "POST",
		QueueConfig: map[interface{}]interface{}{
			"queuename": "test.foo",
		},
	}
	provider := &StaticRepo{Endpoints: []EndpointConfig{epCfg}}
	cfgMgr := NewWatcher(provider, 5)

	// when
	err := cfgMgr.processProviders()
	added := <-cfgMgr.Updates

	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(epCfg, added.Config), fmt.Sprintf("\n   %+v\n!= %+v", epCfg, added.Config))

	provider.Endpoints[0].Uri = "/updated"
	err2 := cfgMgr.processProviders()
	found := <-cfgMgr.Updates

	assert.Nil(t, err2)
	assert.Equal(t, "/updated", found.Config.Uri, "uri should have been updated")
}

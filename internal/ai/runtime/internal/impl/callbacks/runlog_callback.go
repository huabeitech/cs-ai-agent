package callbacks

import (
	"encoding/json"
	"sync"
)

type RuntimeTraceCollector struct {
	mu   sync.Mutex
	Data RuntimeTraceData
}

func NewRuntimeTraceCollector() *RuntimeTraceCollector {
	ret := &RuntimeTraceCollector{}
	ret.Data.Version = "v1"
	ret.Data.Status = "started"
	return ret
}

func (c *RuntimeTraceCollector) Marshal() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	buf, err := json.Marshal(c.Data)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (c *RuntimeTraceCollector) SetTooling(staticToolCodes []string, dynamicToolCodes []string, toolSearchEnabled bool) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Input.StaticToolCodes = append([]string(nil), staticToolCodes...)
	c.Data.Input.DynamicToolCodes = append([]string(nil), dynamicToolCodes...)
	c.Data.Input.ToolSearchEnabled = toolSearchEnabled
}

func (c *RuntimeTraceCollector) AddToolItem(item ToolTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Tools.Count++
	c.Data.Tools.Items = append(c.Data.Tools.Items, item)
}

func (c *RuntimeTraceCollector) AddToolSearchItem(item ToolSearchTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.ToolSearch.Count++
	c.Data.ToolSearch.Items = append(c.Data.ToolSearch.Items, item)
}

func (c *RuntimeTraceCollector) AddGraphToolItem(item GraphToolTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.GraphTools.Count++
	c.Data.GraphTools.Items = append(c.Data.GraphTools.Items, item)
}

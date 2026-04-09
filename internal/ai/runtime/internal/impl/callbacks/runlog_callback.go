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

func (c *RuntimeTraceCollector) AddToolItem(item ToolTraceItem) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Tools.Count++
	c.Data.Tools.Items = append(c.Data.Tools.Items, item)
}

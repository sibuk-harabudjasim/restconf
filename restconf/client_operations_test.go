package restconf

import (
	"fmt"
	"io"
	"testing"

	"io/ioutil"

	"github.com/freeconf/yang/c2"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodes"
	"github.com/freeconf/yang/parser"
)

func Test_ClientOperations(t *testing.T) {
	m, err := parser.LoadModuleFromString(nil, `module x {namespace ""; prefix ""; revision 0;
		container car {			
			container mileage {
				leaf odometer {
					type int32;
				}
				leaf trip {
					type int32;
				}
			}		
			container make {
				leaf model {
					type string;
				}
			}	
		}
}`)
	if err != nil {
		t.Fatal(err)
	}
	support := &testDriverFlowSupport{
		t: t,
	}
	expected := `{"mileage":{"odometer":1000}}`
	support.get = map[string]string{
		"car": expected,
	}
	d := &clientNode{support: support}
	b := node.NewBrowser(m, d.node())
	if actual, err := nodes.WriteJSON(b.Root().Find("car")); err != nil {
		t.Error(err)
	} else {
		c2.AssertEqual(t, expected, actual)
	}

	support.get = map[string]string{
		"car": `{}`,
	}
	expectedEdit := `{"mileage":{"odometer":1001}}`
	edit := nodes.ReadJSON(expectedEdit)
	if err := b.Root().Find("car").UpsertFrom(edit).LastErr; err != nil {
		t.Error(err)
	}
	c2.AssertEqual(t, expectedEdit, support.put["car"])
}

type testDriverFlowSupport struct {
	t    *testing.T
	get  map[string]string
	put  map[string]string
	post map[string]string
}

func (self *testDriverFlowSupport) clientSubscriptions() map[string]*clientSubscription {
	panic("not implemented")
}

func (self *testDriverFlowSupport) clientDo(method string, params string, p *node.Path, payload io.Reader) (node.Node, error) {
	path := p.StringNoModule()
	switch method {
	case "GET":
		in, found := self.get[path]
		if !found {
			return node.ErrorNode{Err: fmt.Errorf("no response for %s", path)}, nil
		}
		return nodes.ReadJSON(in), nil
	case "PUT":
		body, _ := ioutil.ReadAll(payload)
		self.put = map[string]string{
			path: string(body),
		}
	case "POST":
		body, _ := ioutil.ReadAll(payload)
		self.post = map[string]string{
			path: string(body),
		}
	}
	return nil, nil
}

func (self *testDriverFlowSupport) clientSocket() (io.Writer, error) {
	panic("not implemented")
}

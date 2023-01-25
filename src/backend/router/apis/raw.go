package apis

/**
  참조
    https://kubernetes.io/docs/reference/using-api/api-concepts/
    https://github.com/gin-gonic/gin
*/

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kore3lab/dashboard/backend/pkg/app"
	"github.com/kore3lab/dashboard/backend/pkg/config"
	log "github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Get api group list
func GetAPIGroupList(c *gin.Context) {
	g := app.Gin{C: c}

	// instancing dynamic client
	if clientset, err := config.Clusters.NewClientSet(g.C.Param("CLUSTER")); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
	} else {
		if discoveryClient, err := clientset.NewDiscoveryClient(); err != nil {
			g.SendError(err)
		} else {
			if groups, err := discoveryClient.ServerGroups(); err != nil {
				g.SendError(err)
			} else {
				g.Send(http.StatusOK, groups)
			}
		}
	}

}

// Create or Update
func ApplyRaw(c *gin.Context) {
	g := app.Gin{C: c}

	// api client
	if clientset, err := config.Clusters.NewClientSet(g.C.Param("CLUSTER")); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
	} else {
		// invoke POST
		if r, err := clientset.NewKubeExecutor().POST(g.C.Request.Body, g.C.Request.Method == "PUT"); err != nil {
			g.SendMessage(http.StatusBadRequest, err.Error(), err)
		} else {
			g.Send(http.StatusCreated, r)
		}
	}
}

// Delete
func DeleteRaw(c *gin.Context) {
	g := app.Gin{C: c}

	// url parameter validation
	v := []string{"VERSION", "RESOURCE", "NAME"}
	if err := g.ValidateUrl(v); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}

	// instancing dynamic client
	if clientset, err := config.Clusters.NewClientSet(g.C.Param("CLUSTER")); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
	} else {
		// invoke delete
		if client, err := clientset.NewDynamicClint(); err != nil {
			g.SendMessage(http.StatusBadRequest, err.Error(), err)
		} else {
			if err := client.Resource(schema.GroupVersionResource{Group: c.Param("GROUP"), Version: c.Param("VERSION"), Resource: c.Param("RESOURCE")}).Namespace(c.Param("NAMESPACE")).Delete(context.TODO(), c.Param("NAME"), v1.DeleteOptions{}); err != nil {
				g.SendError(err)
			}
		}
	}

}

// Get or List
func GetRaw(c *gin.Context) {
	g := app.Gin{C: c}

	var err error

	ListOptions := v1.ListOptions{}
	query, _ := g.ParseQuery()

	err = v1.Convert_url_Values_To_v1_ListOptions(&query, &ListOptions, nil)
	if err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}
	// instancing dynamic client
	clientset, err := config.Clusters.NewClientSet(g.C.Param("CLUSTER"))
	if err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}

	var r interface{}

	client, err := clientset.NewDynamicClint()
	if err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
	} else {
		executor := client.Resource(schema.GroupVersionResource{Group: c.Param("GROUP"), Version: c.Param("VERSION"), Resource: c.Param("RESOURCE")}).Namespace(c.Param("NAMESPACE"))
		if c.Param("NAME") == "" {
			r, err = executor.List(context.TODO(), ListOptions)
			if err != nil {
				g.SendError(err)
				return
			}
		} else {
			r, err = executor.Get(context.TODO(), c.Param("NAME"), v1.GetOptions{})
			if err != nil {
				if strings.HasSuffix(err.Error(), "not found") {
					g.SendMessage(http.StatusNotFound, err.Error(), err)
				} else {
					g.SendError(err)
				}
				return
			}
		}
		g.Send(http.StatusOK, r)
	}

}

// Patch
func PatchRaw(c *gin.Context) {
	g := app.Gin{C: c}

	// url parameter validation
	v := []string{"VERSION", "RESOURCE", "NAME"}
	if err := g.ValidateUrl(v); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}

	// instancing dynamic client
	if clientset, err := config.Clusters.NewClientSet(g.C.Param("CLUSTER")); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
	} else {
		if client, err := clientset.NewDynamicClint(); err != nil {
			g.SendMessage(http.StatusBadRequest, err.Error(), err)
		} else {
			var r interface{}
			executor := client.Resource(schema.GroupVersionResource{Group: c.Param("GROUP"), Version: c.Param("VERSION"), Resource: c.Param("RESOURCE")}).Namespace(c.Param("NAMESPACE"))
			if payload, err := ioutil.ReadAll(c.Request.Body); err != nil {
				g.SendMessage(http.StatusBadRequest, err.Error(), err)
			} else {
				if r, err = executor.Patch(context.TODO(), c.Param("NAME"), types.PatchType(c.ContentType()), payload, v1.PatchOptions{}); err != nil {
					if strings.HasSuffix(err.Error(), "not found") {
						g.SendMessage(http.StatusNotFound, err.Error(), err)
					} else {
						g.SendError(err)
					}
				} else {
					g.Send(http.StatusOK, r)
				}
			}
		}
	}

}

// Get Pod logs
func GetPodLogs(c *gin.Context) {
	g := app.Gin{C: c}

	var err error

	// url parameter validation
	v := []string{"NAMESPACE", "NAME"}
	if err := g.ValidateUrl(v); err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}

	// instancing dynamic client
	clientset, err := config.Clusters.NewClientSet(g.C.Param("CLUSTER"))
	if err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}

	apiClient, err := clientset.NewKubernetesClient()
	if err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}

	//  log options (with querystring)
	options := coreV1.PodLogOptions{}
	var limitLines = int64(300)
	query, err := g.ParseQuery()
	if err == nil {
		if len(query) > 0 {
			if query["tailLines"] != nil {
				var num1, err1 = strconv.Atoi(query["tailLines"][0])
				if err1 != nil {
					g.SendMessage(http.StatusBadRequest, err.Error(), err)
					return
				}
				limitLines = int64(num1)
			}
			options.TailLines = &limitLines

			if query["sinceTime"] != nil {
				var timestamp v1.Time
				timestamp.UnmarshalQueryParameter(query["sinceTime"][0])
				options.SinceTime = &timestamp
			}

			if query["container"] != nil {
				options.Container = query["container"][0]
			}
			if query["follow"] != nil {
				options.Follow, _ = strconv.ParseBool(query["follow"][0])
			}
			if query["previous"] != nil {
				options.Previous, _ = strconv.ParseBool(query["previous"][0])
			}
			if query["timestamps"] != nil {
				options.Timestamps, _ = strconv.ParseBool(query["timestamps"][0])
			}
		}
	}

	// get a log stream
	req := apiClient.CoreV1().Pods(g.C.Param("NAMESPACE")).GetLogs(g.C.Param("NAME"), &options)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		g.SendMessage(http.StatusBadRequest, err.Error(), err)
		return
	}
	defer stream.Close()

	// read a stream go-routine
	chanStream := make(chan []byte, 10)
	go func() {
		defer close(chanStream)

		for {
			buf := make([]byte, 4096)
			numBytes, err := stream.Read(buf)

			if err != nil {
				if err != io.EOF {
					log.Infof("finished log streaming (cause=%s)", err.Error())
					return
				} else {
					if options.Follow == false {
						log.Debug("log stream is EOF")
						break
					} else {
						time.Sleep(time.Second * 1)
					}
				}
			} else {
				chanStream <- buf[:numBytes]
			}
		}
	}()

	// write stream to client
	g.C.Stream(func(w io.Writer) bool {
		if data, ok := <-chanStream; ok {
			w.Write(data)
			return true
		}
		return false
	})

}

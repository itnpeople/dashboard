package clusters

import (
	"context"
	"encoding/json"
	"testing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kore3lab/dashboard/backend/pkg/client"
)

func TestKubeClusters(t *testing.T) {

	// File based KubeConfig
	t.Log("▶ File based KubeConfig")
	if conf, err := NewKubeClusters(""); err != nil {
		t.Error(err)
	} else {
		if clientset, err := conf.NewClientSet(conf.CurrentCluster); err != nil {
			t.Error(err)
		} else {
			if err := printNamespaces(clientset, t); err != nil {
				t.Error(err)
			} else {
				t.Log("→ OK")
			}
		}
	}

	// ConfigMap based KubeConfig
	//  - prerequisite : in-cluster 환경 9http://itnp.kr/post/client-go)
	//  - kubectl create configmap kore-board-kubeconfig -n default --from-file=config=${HOME}/.kube/config
	//  - kubectl delete configmap kore-board-kubeconfig -n default
	t.Log("▶ ConfigMap based KubeConfig")
	if clusters, err := NewKubeClusters("strategy=configmap,configmap=kore-board-kubeconfig,namespace=default,filename=config"); err != nil {
		t.Error(err)
	} else {
		if client, err := clusters.NewClientSet(clusters.CurrentCluster); err != nil {
			t.Error(err)
		} else {
			if err := printNamespaces(client, t); err != nil {
				t.Error(err)
			} else {
				t.Log("→ OK")
			}
		}
	}

}

func TestKubeClusterAddRemove(t *testing.T) {

	var clusters *KubeClusters
	var err error

	if clusters, err = NewKubeClusters(""); err != nil {
		t.Error(err)
	}
	d := []byte(`{
	    "name": "my-cluster",
	    "cluster": {
	        "server": "https://api-server:6443",
	        "certificate-authority-data": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMxakNDQWI2Z0F3SUJBZ0lKQVBvaTgvYTZqMUIvTUEwR0NTcUdTSWIzRFFFQkN3VUFNQmd4RmpBVUJnTlYKQkFNTURXdDFZbVZ5Ym1WMFpYTXRZMkV3SGhjTk1Ua3hNVEU0TURJek5qQTBXaGNOTWpreE1URTFNREl6TmpBMApXakFZTVJZd0ZBWURWUVFEREExcmRXSmxjbTVsZEdWekxXTmhNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DCkFROEFNSUlCQ2dLQ0FRRUEydjZMbE9UVXV4Y0VQV1BETTRiNVROdHNBU2I0VGFObTFpVmdid093MXIxK3lBUGsKL09lMFFLQXRGejRDb3ZkcEZwMTh3YkxlcDY0cFdDZ3VUcytLTGlqSWtpenRQcG1KeWplUUdaSlJwY2J0dFNuMgp5MXZCNkh1UDBnKzZ0bjVzWSs5STFMbXN0ajVJZ255VWxHUnkzU1I4d0luM3NUdW5qWWhOL0QvVzBDQTRhVG9vCjNTV1R4d011RUVjc1hCeVYwWG8rTzErZVNpSitUWWQ5VTFPd3dUU1gvSnVod1dOZHhRejZISDd0N1pOOU4xWU0KT05PeFdaY29UanB5R0pycko5WmZzWHhDdkErbDVsQnQyZHNSMC8vRVlVR2FLRXVnZ0R2am1tOHl6dkw1MFRYSgowVW1lMTRMMzBrbTdLQjJJMVRJNlY1UmJWbTJIRkROcnpYNTVld0lEQVFBQm95TXdJVEFQQmdOVkhSTUJBZjhFCkJUQURBUUgvTUE0R0ExVWREd0VCL3dRRUF3SUNwREFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBMThZYzk4b0gKMTNFZFVDN2tCT0VNMm84NU96M1BVZkVtZG40OTRNQmdSR1FodytYWXAzOVAyWVVhT09SbDFQQ2FaSy82TjZsRApLcVFXNUxCYjhxaExQejZ5Rm5YQ2t6ODBtM2JUdWc1aEdHVTFHV3B0N1E1N0EzS2VWbEttOVRVR3RSdTZ3SGE1CkFNYlh2SDc4cHFBQzhjTFV3OTlQMjNzTWpxMDlVQWpSV0tVTmxmbmpTQ2NUTnVNRExsNWFIV2ttMW1VRjQwWDEKRDREWEk0clhTY3p2TlYzdEN1YzBPanpvRjZ4aHBiT01ucGU1QjFKZlpyeCtJZi9sRzJGM3BPN0ZnNkxNOXhJVAo5a1gyQ0x0SnhVYXhiZVplZWRUeWVHbjBFV1hDMGpubGNHTkpHdXcrbk9ndVNKNmJpYWt6OFlTaTdKN21HSFlXCjJlWGNzOVJSSUg3SWVRPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	    },
	    "user": {
	        "client-certificate-data": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM5VENDQWQyZ0F3SUJBZ0lJS0hzQzRjQzlKbkl3RFFZSktvWklodmNOQVFFTEJRQXdHREVXTUJRR0ExVUUKQXd3TmEzVmlaWEp1WlhSbGN5MWpZVEFlRncweE9URXhNVGd3TWpNMk1EUmFGdzB5TVRFd01qWXdOakUxTVRaYQpNRFF4RnpBVkJnTlZCQW9URG5ONWMzUmxiVHB0WVhOMFpYSnpNUmt3RndZRFZRUURFeEJyZFdKbGNtNWxkR1Z6CkxXRmtiV2x1TUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUEydzllMU91WGM5VXkKZGdYUnhSNTFXR3dMMnNHdEJFdHU1czRqb3d1Y1RJZTFmWjBpZjJEZERadDVPK0NNazhJL1ZqeW9LSUtNR09LSQpiTmdSNUp1NmpNQ29PN1JsV3pxT29UaFE0emNtSmMrbTZBYkF0L0IwTVpxL3g0ZlFjR0ZSc09hZlFBdGxhWjlwCkIwTjI4Rm4zalBJZE9wWmx5ZFY5eUp2L2JiQTBtczN1YU5udHpkS2pZOWJ2Ti9oYVZ0NmEyYlViSWhzSzJmNlIKRUFnaURwZlhQd0VBd3NZbm16U2ZJdy9URTA1VzNicmZxdSsvdWxPbFBxdHM0YlBOR3pLb1hjaERLcUlmZnBBMApTQWZCeEJodUw5N1VLTkNIU1lidGtSeFpCM1VGeGs1aXdxRFJTakRaYzZBTm0xOVh6L05NN0hwdk02UnFMM0RyCitKZ25oaGc4NlFJREFRQUJveWN3SlRBT0JnTlZIUThCQWY4RUJBTUNCYUF3RXdZRFZSMGxCQXd3Q2dZSUt3WUIKQlFVSEF3SXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBQ0w4R3hUZTExcGZHUnhsanpWQW9UTDdBL29JM2xmNwozbi9LRE1ZQkdjNmZ2Snhuay9QeGFnd2p4QUtkeVVjaHdIOEg2OXloVU45MHZIM2hZNStmRENTYWtpaUlWSTZ4ClFZT1RZVzJyYXJwa2dOZEFBNi9CRmlndVJOajJqYkVxenFvOGl1UlhrNGxieS9VbVkrT25IVTVtL3NKQXVkWEIKeUs5RnVVdWNyUUhtR1Z4cThXMnFXNDRUU05JY2lFdVpycXlQdW55QUNjNmRsc21MSWkrNWFUbC9SYjRqQTFQbApUaWQrOWFhSGJISDg3emRlanZqajhBc1pQVlN2azBDRy9TbnRRSnkwMHhlbU1rcTJ0eDdJdmpHVkdtdzRUeThOCnFPVGhnTzhCQ3ZDdys5RzRWVHBUei9jckgvaVhXWEZadVN6L3JoSjdLNmF4Z1V4MmtOc1lnZG89Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K",
	        "client-key-data": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBMnc5ZTFPdVhjOVV5ZGdYUnhSNTFXR3dMMnNHdEJFdHU1czRqb3d1Y1RJZTFmWjBpCmYyRGREWnQ1TytDTWs4SS9WanlvS0lLTUdPS0liTmdSNUp1NmpNQ29PN1JsV3pxT29UaFE0emNtSmMrbTZBYkEKdC9CME1acS94NGZRY0dGUnNPYWZRQXRsYVo5cEIwTjI4Rm4zalBJZE9wWmx5ZFY5eUp2L2JiQTBtczN1YU5udAp6ZEtqWTlidk4vaGFWdDZhMmJVYkloc0syZjZSRUFnaURwZlhQd0VBd3NZbm16U2ZJdy9URTA1VzNicmZxdSsvCnVsT2xQcXRzNGJQTkd6S29YY2hES3FJZmZwQTBTQWZCeEJodUw5N1VLTkNIU1lidGtSeFpCM1VGeGs1aXdxRFIKU2pEWmM2QU5tMTlYei9OTTdIcHZNNlJxTDNEcitKZ25oaGc4NlFJREFRQUJBb0lCQVFDTjRLQ2M2cEZHOWxnZQpWSnFPUHJIbHVPVGNvLys0L2xvdnBtY3lYSHk0bkZTUnJNb0JFZTFadU02R1YyTDArQ1FwYUZQSkdQUS8wY2htCkpuTkFTeFJCd1MyMHJadlB3RmRNVjdzYnprWW95eHJnd0M3bGN4anVYN25DTTFadTBya2tCOW93a3JEMS9jYjYKYTFtSFJkMnRMY3A4Zlpnalp1QjJvNEtGdWsvM3lpSFVncHNXOXQ5MEVuV1VrbkxoSEQyckpsOElLRHpVTEViVgorRlkya3ZzRnNlU0tMYXlYVVRmSG5VVkgyczR6R2thYWVXQWhvSkUyTHpYaHdkeGV4aGFmNHBiUkMrSTJtaHRECng0Q2YvT3o5ZC81SjBHbGk3Zlo1Q2ZpMDdsd09DVEJ2NGNIcVU3cjZlbm5BVUVGNHhTUVc5dytGL2lDdG9HemcKVTAvSmg4ZmhBb0dCQU94emkrZE9OdTliRUI2MGZ2enpmYy8xVXJHNFFUKzdwL3FoVXpLSllIOUYySURxYjVpRwpKcUNKUU9STW9WcEpNSDVkSXB3QlBCeXA0YkV1NU8rSGdjT0o4SmhZYlBLZnREeTU3dHpPanE0bXlieGxObjhMCjFOdXU0Y3NrNmcwZWJ1QnZveGROMkUvNTc0TE5GVzcwR1dXaDJLdmhpL2xSRU9JZkZlUkh1Nk1kQW9HQkFPMHIKdlYwVnVZNGQwM1UrUzg3NFk5b0l1bWZnM05iYVRMb0F3RVh1Qi9jTVJZVWoyT3l1TFB6OTBzTmVhNFlVZFpNQwpLNWl2bURqWUpFOHExNmlDdW4vSDlDUnZzS05rZ2kvKzQxS29WRmdVVjcwdFBXWGNjOXNOMnc3UWQ0dUFOMjNQCnlaSHAraHJUYWJYalRTVEVONTkrM2Z6TnZLMmo4MHNFcDZTd3hxczlBb0dBSHgxVlc2cS9MK0FienU2UmgxZkQKUm9wUngzRW5wT3RjdjI1Yk5Gcy9oMy81Ylgxc0VmWVZQeXJRanpwR1FVdEFSbUNiSFV4TVRMbE9LYkt5RFpNWApVRlBtaFNXZHNJK3plQW8vbEc3Wjk3REMremVXWkVGNlVTNUNLQ2xEWTFhTjRKclFLMURqRmlNZGtXakxXVDVsCjJTbmpDVHMwNENuNnZzYTRhc0hGdjBFQ2dZRUEzd1IwU21XMVdGZmZrYTRFcHhpVy9GMmN1eldOTkZPT05wR2kKTzUrNnlhbzJiUjNxUzVUMUpPaWhHYWxkdm5UYW9tUTJEcHQvcm1SQXNGais5YXdJSjBRazVXWkpXVHVYMS8zOApVS3VNdEU1Y3VyMGhzUGo5MEl4VTRyZFEwbEs4ekh2SmRYWHBBdlN0d0tWKzB3WFhzQmtpTVNoZE5ZS25zbkVzCkd3ZEhxWmtDZ1lFQTJFaTRXdGh5MlBVUkRHbHViWHdHSXFhWmRpMlMyVGYreUJyQ2dWRFQ1T3hNaFZ2dDVoN28KQ2syYithT0JCaGIvdnBvVFQya0ROUEZCWUxMYTRpMkRJRGN2emkzRHVUcDkxUWVqN2tJYi9oMlQ2aVAyUkprcAp3YlM3aFFPUVk1SWRpNFRrWGliUk93c3ArM25QdHd4Q0hPblREZVEwZHV3UHNtZmFvS2VLRkdrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	    }
	}`)

	var params KubeConfig
	if err := json.Unmarshal(d, &params); err != nil {
		t.Error(err)
	} else {
		if err = clusters.AddCluster(&params); err != nil {
			t.Error(err)
		} else {
			t.Log("▶ Add a Clusater → OK")
			if err = clusters.RemoveCluster(params.Name); err != nil {
				t.Error(err)
			} else {
				t.Log("▶ Remove a Clusater → OK")
			}
		}
	}

}

func printNamespaces(clientset *client.ClientSet, t *testing.T) error {

	if client, err := clientset.NewKubernetesClient(); err != nil {
		return err
	} else if list, err := client.CoreV1().Namespaces().List(context.TODO(), metaV1.ListOptions{}); err != nil {
		return err
	} else {
		t.Logf("  namespace (%d ea)", len(list.Items))
		for i, ns := range list.Items {
			t.Logf("    %d : %s", i+1, ns.ObjectMeta.Name)
		}
	}
	return nil
}

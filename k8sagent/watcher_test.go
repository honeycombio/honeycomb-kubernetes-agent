package k8sagent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestPodWatcherListsPodsWithoutStreamingListSupport(t *testing.T) {
	// An apiserver that just serves a list and an idle watch. The informer has no
	// stop method, so the server is left running for the duration of the test binary.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("watch") == "true" {
			w.WriteHeader(200)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return
		}
		json.NewEncoder(w).Encode(v1.PodList{
			TypeMeta: metav1.TypeMeta{Kind: "PodList", APIVersion: "v1"},
			ListMeta: metav1.ListMeta{ResourceVersion: "5"},
			Items:    []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "a", UID: "uid-a"}}},
		})
	}))

	cs := kubernetes.NewForConfigOrDie(&rest.Config{Host: srv.URL})
	w := NewPodWatcher("", "app=nginx", "", cs.CoreV1())
	select {
	case p := <-w.Pods():
		if p.Name != "a" {
			t.Fatalf("unexpected pod %s", p.Name)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no pod within 5s")
	}
}

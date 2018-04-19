package kube

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bjaglin/multiplexio"
	pb "github.com/datacol-io/datacol/api/models"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func LogStreamReq(c *kubernetes.Clientset, w io.Writer, ns, app string, opts pb.LogStreamOptions) error {
	pods, err := GetAllRunningPods(c, ns, app)
	if err != nil {
		return err
	}

	log.Debugf("Got %d pods for app=%s", len(pods), app)

	//TODO: consider using https://github.com/djherbis/stream for reading multiple streams
	var sources []multiplexio.Source
	type reqQueue struct {
		name    string
		request *rest.Request
	}

	var requests []reqQueue

	for _, pod := range pods {
		if opts.Proctype != "" && opts.Proctype != pod.ObjectMeta.Labels[typeLabel] {
			continue
		}

		// Don't collect logs from ephemeral pods
		if proctype := pod.ObjectMeta.Labels[typeLabel]; proctype == runProcessKind {
			continue
		}

		name := pod.Name
		req := c.Core().RESTClient().Get().
			Namespace(ns).
			Name(name).
			Resource("pods").
			SubResource("log").
			Param("follow", strconv.FormatBool(opts.Follow)).
			Param("tailLines", opts.TailLines)

		var cntName string
		if len(pod.Spec.Containers) > 0 {
			cntName = pod.Spec.Containers[0].Name
		}

		req = req.Param("container", cntName)
		log.Debugf("will stream logs from %v/%s", name, cntName)

		if opts.Since > 0 {
			sec := int64(math.Ceil(float64(opts.Since) / float64(time.Second)))
			req = req.Param("sinceSeconds", strconv.FormatInt(sec, 10))
		}

		requests = append(requests, reqQueue{name: name, request: req})
	}

	var wg sync.WaitGroup

	for _, rq := range requests {
		wg.Add(1)

		go func(rq reqQueue) {
			defer wg.Done()
			r, err := rq.request.Stream()
			if err != nil {
				log.Errorf("creating log stream: %v", err)
			}

			prefix := fmt.Sprintf("[%s] ", strings.TrimPrefix(rq.name, app+"-"))
			sources = append(sources, multiplexio.Source{
				Reader: r,
				Write: func(dest io.Writer, token []byte) (int, error) {
					return multiplexio.WriteNewLine(dest, append([]byte(prefix), token...))
				},
			})
		}(rq)
	}

	log.Infof("waiting for stream handlers ...")
	wg.Wait()
	log.Debugf("Done. Got %d streams", len(sources))

	_, err = io.Copy(w, multiplexio.NewReader(multiplexio.Options{}, sources...))
	return err
}

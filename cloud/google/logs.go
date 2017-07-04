package google

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"math"
	"strconv"
	"time"
)

func (g *GCPCloud) LogStream(app string, opts pb.LogStreamOptions) (*bufio.Reader, func() error, error) {
	ns := g.DeploymentName
	c, err := getKubeClientset(ns)
	if err != nil {
		return nil, nil, err
	}

	pod, err := runningPods(ns, app, c)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Getting logs from pod %s", pod)

	req := c.Core().RESTClient().Get().
		Namespace(ns).
		Name(pod).
		Resource("pods").
		SubResource("log").
		Param("container", app).
		Param("follow", strconv.FormatBool(opts.Follow))

	if opts.Since > 0 {
		sec := int64(math.Ceil(float64(opts.Since) / float64(time.Second)))
		req = req.Param("sinceSeconds", strconv.FormatInt(sec, 10))
	}

	rc, err := req.Stream()
	if err != nil {
		return nil, nil, err
	}

	return bufio.NewReader(rc), rc.Close, nil
}

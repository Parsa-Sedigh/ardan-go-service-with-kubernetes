package v1

import (
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/data/dbtest"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/docker"
	"testing"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error

	/* Note: We only need one DB instance to run all the route-level tests as opposed to unit tests that each has it's own DB instance. */
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)

		return
	}

	defer dbtest.StopDB(c)

	m.Run()
}

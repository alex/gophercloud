// +build acceptance blockstorage

package v1

import (
	"os"
	"testing"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/pagination"
)

func newClient() (*gophercloud.ServiceClient, error) {
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	client, err := openstack.AuthenticatedClient(ao)
	if err != nil {
		return nil, err
	}

	return openstack.NewBlockStorageV1(client, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}

func TestVolumes(t *testing.T) {
	client, err := newClient()
	if err != nil {
		t.Fatalf("Failed to create Block Storage v1 client: %v", err)
	}

	cv, err := volumes.Create(client, &volumes.CreateOpts{
		Size: 1,
		Name: "gophercloud-test-volume",
	}).Extract()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = volumes.WaitForStatus(client, cv.ID, "available", 60)
		if err != nil {
			t.Error(err)
		}
		res = volumes.Delete(client, cv.ID)
		if res.Err != nil {
			t.Error(err)
			return
		}
	}()

	_, err = volumes.Update(client, cv.ID, &volumes.UpdateOpts{
		Name: "gophercloud-updated-volume",
	}).Extract()
	if err != nil {
		t.Error(err)
		return
	}

	v, err := volumes.Get(client, cv.ID).Extract()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Got volume: %+v\n", v)

	if v.Name != "gophercloud-updated-volume" {
		t.Errorf("Unable to update volume: Expected name: gophercloud-updated-volume\nActual name: %s", v.Name)
	}

	err = volumes.List(client, &volumes.ListOpts{Name: "gophercloud-updated-volume"}).EachPage(func(page pagination.Page) (bool, error) {
		vols, err := volumes.ExtractVolumes(page)
		if len(vols) != 1 {
			t.Errorf("Expected 1 volume, got %d", len(vols))
		}
		return true, err
	})
	if err != nil {
		t.Errorf("Error listing volumes: %v", err)
	}
}

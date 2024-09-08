package provider

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
)

type providerLinode struct {
	clients *clientmap.ClientMap
}

func NewProviderLinode() (*providerLinode, error) {
	pl := new(providerLinode)

	clients, err := pl.newClients()
	if err != nil {
		return pl, err
	}
	pl.clients = clients
	return pl, nil
}

func (pl *providerLinode) getRegionClient(region string) *s3.Client {
	return pl.clients.Get(region, false)
}

func (pl *providerLinode) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = pl.Name()
	exists, region, err := bucketExists(pl.clients, b)
	if err != nil {
		return b, err
	}
	if exists {
		b.Exists = bucket.BucketExists
		b.Region = region
	} else {
		b.Exists = bucket.BucketNotExist
	}

	return b, nil
}

func (pl *providerLinode) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}

	client := pl.getRegionClient(b.Region)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func (pl *providerLinode) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(ProviderRegions[pl.Name()]))
	for _, r := range ProviderRegions[pl.Name()] {
		client, err := newNonAWSClient(pl, fmt.Sprintf("https://%s.linodeobjects.com", r))
		if err != nil {
			return nil, err
		}
		clients.Set(r, false, client)
	}

	return clients, nil
}

func (pl *providerLinode) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	client := pl.getRegionClient(b.Region)
	return checkPermissions(client, b, doDestructiveChecks)
}

func (*providerLinode) Insecure() bool {
	return false
}

func (*providerLinode) Name() string {
	return "linode"
}

func (*providerLinode) AddressStyle() int {
	return VirtualHostStyle
}

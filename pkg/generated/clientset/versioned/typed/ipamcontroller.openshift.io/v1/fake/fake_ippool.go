// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	ipamcontrolleropenshiftiov1 "github.com/openshift-splat-team/machine-ipam-controller/pkg/apis/ipamcontroller.openshift.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeIPPools implements IPPoolInterface
type FakeIPPools struct {
	Fake *FakeIpamcontrollerV1
	ns   string
}

var ippoolsResource = schema.GroupVersionResource{Group: "ipamcontroller.openshift.io", Version: "v1", Resource: "ippools"}

var ippoolsKind = schema.GroupVersionKind{Group: "ipamcontroller.openshift.io", Version: "v1", Kind: "IPPool"}

// Get takes name of the iPPool, and returns the corresponding iPPool object, and an error if there is any.
func (c *FakeIPPools) Get(ctx context.Context, name string, options v1.GetOptions) (result *ipamcontrolleropenshiftiov1.IPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ippoolsResource, c.ns, name), &ipamcontrolleropenshiftiov1.IPPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ipamcontrolleropenshiftiov1.IPPool), err
}

// List takes label and field selectors, and returns the list of IPPools that match those selectors.
func (c *FakeIPPools) List(ctx context.Context, opts v1.ListOptions) (result *ipamcontrolleropenshiftiov1.IPPoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ippoolsResource, ippoolsKind, c.ns, opts), &ipamcontrolleropenshiftiov1.IPPoolList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &ipamcontrolleropenshiftiov1.IPPoolList{ListMeta: obj.(*ipamcontrolleropenshiftiov1.IPPoolList).ListMeta}
	for _, item := range obj.(*ipamcontrolleropenshiftiov1.IPPoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested iPPools.
func (c *FakeIPPools) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ippoolsResource, c.ns, opts))

}

// Create takes the representation of a iPPool and creates it.  Returns the server's representation of the iPPool, and an error, if there is any.
func (c *FakeIPPools) Create(ctx context.Context, iPPool *ipamcontrolleropenshiftiov1.IPPool, opts v1.CreateOptions) (result *ipamcontrolleropenshiftiov1.IPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ippoolsResource, c.ns, iPPool), &ipamcontrolleropenshiftiov1.IPPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ipamcontrolleropenshiftiov1.IPPool), err
}

// Update takes the representation of a iPPool and updates it. Returns the server's representation of the iPPool, and an error, if there is any.
func (c *FakeIPPools) Update(ctx context.Context, iPPool *ipamcontrolleropenshiftiov1.IPPool, opts v1.UpdateOptions) (result *ipamcontrolleropenshiftiov1.IPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ippoolsResource, c.ns, iPPool), &ipamcontrolleropenshiftiov1.IPPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ipamcontrolleropenshiftiov1.IPPool), err
}

// Delete takes name of the iPPool and deletes it. Returns an error if one occurs.
func (c *FakeIPPools) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(ippoolsResource, c.ns, name, opts), &ipamcontrolleropenshiftiov1.IPPool{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIPPools) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ippoolsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &ipamcontrolleropenshiftiov1.IPPoolList{})
	return err
}

// Patch applies the patch and returns the patched iPPool.
func (c *FakeIPPools) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *ipamcontrolleropenshiftiov1.IPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ippoolsResource, c.ns, name, pt, data, subresources...), &ipamcontrolleropenshiftiov1.IPPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ipamcontrolleropenshiftiov1.IPPool), err
}

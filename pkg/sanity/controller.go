/*
Copyright 2017 Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sanity

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	context "golang.org/x/net/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func verifyVolumeInfo(v *csi.VolumeInfo) {
	Expect(v).NotTo(BeNil())
	Expect(v.GetId()).NotTo(BeEmpty())
}

func isCapabilitySupported(
	c csi.ControllerClient,
	capType csi.ControllerServiceCapability_RPC_Type,
) bool {

	caps, err := c.ControllerGetCapabilities(
		context.Background(),
		&csi.ControllerGetCapabilitiesRequest{
			Version: csiClientVersion,
		})
	Expect(err).NotTo(HaveOccurred())
	Expect(caps).NotTo(BeNil())
	Expect(caps.GetCapabilities()).NotTo(BeNil())

	for _, cap := range caps.GetCapabilities() {
		Expect(cap.GetRpc()).NotTo(BeNil())
		if cap.GetRpc().GetType() == capType {
			return true
		}
	}
	return false
}

var _ = Describe("ControllerGetCapabilities [Controller Server]", func() {
	var (
		c csi.ControllerClient
	)

	BeforeEach(func() {
		c = csi.NewControllerClient(conn)
	})

	It("should fail when no version is provided", func() {
		_, err := c.ControllerGetCapabilities(
			context.Background(),
			&csi.ControllerGetCapabilitiesRequest{})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should return appropriate capabilities", func() {
		caps, err := c.ControllerGetCapabilities(
			context.Background(),
			&csi.ControllerGetCapabilitiesRequest{
				Version: csiClientVersion,
			})

		By("checking successful response")
		Expect(err).NotTo(HaveOccurred())
		Expect(caps).NotTo(BeNil())
		Expect(caps.GetCapabilities()).NotTo(BeNil())

		for _, cap := range caps.GetCapabilities() {
			Expect(cap.GetRpc()).NotTo(BeNil())

			switch cap.GetRpc().GetType() {
			case csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME:
			case csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME:
			case csi.ControllerServiceCapability_RPC_LIST_VOLUMES:
			case csi.ControllerServiceCapability_RPC_GET_CAPACITY:
			default:
				Fail(fmt.Sprintf("Unknown capability: %v\n", cap.GetRpc().GetType()))
			}
		}
	})
})

var _ = Describe("GetCapacity [Controller Server]", func() {
	var (
		c csi.ControllerClient
	)

	BeforeEach(func() {
		c = csi.NewControllerClient(conn)

		if !isCapabilitySupported(c, csi.ControllerServiceCapability_RPC_GET_CAPACITY) {
			Skip("GetCapacity not supported")
		}
	})

	It("should fail when no version is provided", func() {

		By("failing when there is no version")
		_, err := c.GetCapacity(
			context.Background(),
			&csi.GetCapacityRequest{})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should return capacity (no optional values added)", func() {
		_, err := c.GetCapacity(
			context.Background(),
			&csi.GetCapacityRequest{
				Version: csiClientVersion,
			})
		Expect(err).NotTo(HaveOccurred())

		// Since capacity is uint64 we will not be checking it
		// The value of zero is a possible value.
	})
})

var _ = Describe("ListVolumes [Controller Server]", func() {
	var (
		c csi.ControllerClient
	)

	BeforeEach(func() {
		c = csi.NewControllerClient(conn)

		if !isCapabilitySupported(c, csi.ControllerServiceCapability_RPC_LIST_VOLUMES) {
			Skip("ListVolumes not supported")
		}
	})

	It("should fail when no version is provided", func() {

		By("failing when there is no version")
		_, err := c.ListVolumes(
			context.Background(),
			&csi.ListVolumesRequest{})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should return appropriate values (no optional values added)", func() {
		vols, err := c.ListVolumes(
			context.Background(),
			&csi.ListVolumesRequest{
				Version: csiClientVersion,
			})
		Expect(err).NotTo(HaveOccurred())
		Expect(vols).NotTo(BeNil())
		Expect(vols.GetEntries()).NotTo(BeNil())

		for _, vol := range vols.GetEntries() {
			verifyVolumeInfo(vol.GetVolumeInfo())
		}
	})

	// TODO: Add test to test for tokens

	// TODO: Add test which checks list of volume is there when created,
	//       and not there when deleted.
})

var _ = Describe("CreateVolume [Controller Server]", func() {
	var (
		c csi.ControllerClient
	)

	BeforeEach(func() {
		c = csi.NewControllerClient(conn)

		if !isCapabilitySupported(c, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
			Skip("CreateVolume not supported")
		}
	})

	It("should fail when no version is provided", func() {

		_, err := c.CreateVolume(
			context.Background(),
			&csi.CreateVolumeRequest{})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should fail when no name is provided", func() {

		_, err := c.CreateVolume(
			context.Background(),
			&csi.CreateVolumeRequest{
				Version: csiClientVersion,
			})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should fail when no volume capabilities are provided", func() {

		_, err := c.CreateVolume(
			context.Background(),
			&csi.CreateVolumeRequest{
				Version: csiClientVersion,
				Name:    "name",
			})
		Expect(err).To(HaveOccurred())

		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(Equal(codes.InvalidArgument))
	})

	It("should return appropriate values SingleNodeWriter NoCapacity Type:Mount", func() {
		name := "sanity"
		vol, err := c.CreateVolume(
			context.Background(),
			&csi.CreateVolumeRequest{
				Version: csiClientVersion,
				Name:    name,
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			})
		Expect(err).NotTo(HaveOccurred())
		Expect(vol).NotTo(BeNil())
		Expect(vol.GetVolumeInfo()).NotTo(BeNil())
		Expect(vol.GetVolumeInfo().GetId()).NotTo(BeEmpty())
	})

	// Pending fix in mock file
	It("[MOCKERRORS] should return appropriate values SingleNodeWriter WithCapacity 1Gi Type:Mount", func() {
		name := "sanity"
		size := uint64(1 * 1024 * 1024 * 1024)
		vol, err := c.CreateVolume(
			context.Background(),
			&csi.CreateVolumeRequest{
				Version: csiClientVersion,
				Name:    name,
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: size,
				},
			})
		Expect(err).NotTo(HaveOccurred())
		Expect(vol).NotTo(BeNil())
		Expect(vol.GetVolumeInfo()).NotTo(BeNil())
		Expect(vol.GetVolumeInfo().GetId()).NotTo(BeEmpty())
		Expect(vol.GetVolumeInfo().GetCapacityBytes()).To(Equal(size))
	})
})
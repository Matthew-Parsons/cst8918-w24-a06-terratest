package test

import (
	"testing"
	"strings"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "8a667ada-0a9e-458e-9674-f4fabbe9aed3"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "MatthewParsons",
		},
	}

	// Clean up resources
	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of output variable
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name")

	// Confirm VM exists
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// Confirm NIC exists
	assert.True(t, azure.NetworkInterfaceExists(t, nicName, resourceGroupName, subscriptionID))

	// Confirm NIC is attached to the VM
	nicIDs, err := azure.GetVirtualMachineNicsE(vmName, resourceGroupName, subscriptionID)
	assert.NoError(t, err)

	found := false
	for _, id := range nicIDs {
		if strings.Contains(id, nicName) {
			found = true
		}
	}
	assert.True(t, found, "NIC %s is not attached to VM %s", nicName, vmName)

	// Fetch VM object for image inspection
	vm, err := azure.GetVirtualMachineE(vmName, resourceGroupName, subscriptionID)
	assert.NoError(t, err)
	
	// Confirm VM uses Ubuntu publisher/offer
	image := vm.StorageProfile.ImageReference
	assert.Equal(t, "Canonical", *image.Publisher)
	assert.Equal(t, "UbuntuServer", *image.Offer)

	// Confirm Ubuntu version contains expected SKU
	assert.Contains(t, *image.Sku, "22_04-lts-gen2")
}

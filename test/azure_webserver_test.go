package test

import (
	"testing"

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
	nic := azure.GetNetworkInterface(t, nicName, resourceGroup, subscriptionID)
	assert.Equal(t, nicName, *nic.Name)

	// Confirm NIC is attached to the VM
	vm := azure.GetVirtualMachine(t, vmName, resourceGroup, subscriptionID)

	found := false
	for _, vmNic := range *vm.NetworkProfile.NetworkInterfaces {

		if strings.Contains(*vmNic.ID, nicName) {
			found = true
		}
	}

	assert.True(t, found, "NIC is not attached to VM")

	// Confirm the VM uses Ubuntu
	image := vm.StorageProfile.ImageReference

	assert.Equal(t, "Canonical", *image.Publisher)
	assert.Equal(t, "UbuntuServer", *image.Offer)

	// Confirm Ubuntu version
	assert.Contains(t, *image.Sku, "22_04-lts-gen2")

}

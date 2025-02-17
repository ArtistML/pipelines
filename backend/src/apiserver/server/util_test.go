// Copyright 2018-2022 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	apiv1beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"
	apiv2beta1 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/kubeflow/pipelines/backend/src/apiserver/resource"
	"github.com/kubeflow/pipelines/backend/src/common/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGetPipelineName_QueryStringNotEmpty(t *testing.T) {
	pipelineName, err := GetPipelineName("pipeline%20one", "file one")
	assert.Nil(t, err)
	assert.Equal(t, "pipeline one", pipelineName)
}

func TestGetPipelineName(t *testing.T) {
	pipelineName, err := GetPipelineName("", "file one")
	assert.Nil(t, err)
	assert.Equal(t, "file one", pipelineName)
}

func TestGetPipelineName_InvalidQueryString(t *testing.T) {
	_, err := GetPipelineName("pipeline!$%one", "file one")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestGetPipelineName_NameTooLong(t *testing.T) {
	_, err := GetPipelineName("",
		"this is a loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooog name")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "name too long")
}

func TestLoadFile(t *testing.T) {
	file := "12345"
	bytes, err := loadFile(strings.NewReader(file), 5)
	assert.Nil(t, err)
	assert.Equal(t, []byte(file), bytes)
}

func TestLoadFile_ExceedSizeLimit(t *testing.T) {
	file := "12345"
	_, err := loadFile(strings.NewReader(file), 4)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "File size too large")
}

func TestDecompressPipelineTarball(t *testing.T) {
	tarballByte, _ := ioutil.ReadFile("test/arguments_tarball/arguments.tar.gz")
	pipelineFile, err := DecompressPipelineTarball(tarballByte)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestDecompressPipelineTarball_MalformattedTarball(t *testing.T) {
	tarballByte, _ := ioutil.ReadFile("test/malformatted_tarball.tar.gz")
	_, err := DecompressPipelineTarball(tarballByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Not a valid tarball file")
}

func TestDecompressPipelineTarball_NonYamlTarball(t *testing.T) {
	tarballByte, _ := ioutil.ReadFile("test/non_yaml_tarball/non_yaml_tarball.tar.gz")
	_, err := DecompressPipelineTarball(tarballByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Expecting a pipeline.yaml file inside the tarball")
}

func TestDecompressPipelineTarball_EmptyTarball(t *testing.T) {
	tarballByte, _ := ioutil.ReadFile("test/empty_tarball/empty.tar.gz")
	_, err := DecompressPipelineTarball(tarballByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Not a valid tarball file")
}

func TestDecompressPipelineZip(t *testing.T) {
	zipByte, _ := ioutil.ReadFile("test/arguments_zip/arguments-parameters.zip")
	pipelineFile, err := DecompressPipelineZip(zipByte)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestDecompressPipelineZip_MalformattedZip(t *testing.T) {
	zipByte, _ := ioutil.ReadFile("test/malformatted_zip.zip")
	_, err := DecompressPipelineZip(zipByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Not a valid zip file")
}

func TestDecompressPipelineZip_MalformedZip2(t *testing.T) {
	zipByte, _ := ioutil.ReadFile("test/malformed_zip2.zip")
	_, err := DecompressPipelineZip(zipByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Not a valid zip file")
}

func TestDecompressPipelineZip_NonYamlZip(t *testing.T) {
	zipByte, _ := ioutil.ReadFile("test/non_yaml_zip/non_yaml_file.zip")
	_, err := DecompressPipelineZip(zipByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Expecting a pipeline.yaml file inside the zip")
}

func TestDecompressPipelineZip_EmptyZip(t *testing.T) {
	zipByte, _ := ioutil.ReadFile("test/empty_tarball/empty.zip")
	_, err := DecompressPipelineZip(zipByte)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Not a valid zip file")
}

func TestReadPipelineFile_YAML(t *testing.T) {
	file, _ := os.Open("test/arguments-parameters.yaml")
	fileBytes, err := ReadPipelineFile("arguments-parameters.yaml", file, MaxFileLength)
	assert.Nil(t, err)

	expectedFileBytes, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedFileBytes, fileBytes)
}

func TestReadPipelineFile_JSON(t *testing.T) {
	file, _ := os.Open("test/v2-hello-world.json")
	fileBytes, err := ReadPipelineFile("v2-hello-world.json", file, MaxFileLength)
	assert.Nil(t, err)

	expectedFileBytes, _ := ioutil.ReadFile("test/v2-hello-world.json")
	assert.Equal(t, expectedFileBytes, fileBytes)
}

func TestReadPipelineFile_Zip(t *testing.T) {
	file, _ := os.Open("test/arguments_zip/arguments-parameters.zip")
	pipelineFile, err := ReadPipelineFile("arguments-parameters.zip", file, MaxFileLength)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestReadPipelineFile_Zip_AnyExtension(t *testing.T) {
	file, _ := os.Open("test/arguments_zip/arguments-parameters.zip")
	pipelineFile, err := ReadPipelineFile("arguments-parameters.pipeline", file, MaxFileLength)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestReadPipelineFile_MultifileZip(t *testing.T) {
	file, _ := os.Open("test/pipeline_plus_component/pipeline_plus_component.zip")
	pipelineFile, err := ReadPipelineFile("pipeline_plus_component.ai-hub-package", file, MaxFileLength)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/pipeline_plus_component/pipeline.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestReadPipelineFile_Tarball(t *testing.T) {
	file, _ := os.Open("test/arguments_tarball/arguments.tar.gz")
	pipelineFile, err := ReadPipelineFile("arguments.tar.gz", file, MaxFileLength)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestReadPipelineFile_Tarball_AnyExtension(t *testing.T) {
	file, _ := os.Open("test/arguments_tarball/arguments.tar.gz")
	pipelineFile, err := ReadPipelineFile("arguments.pipeline", file, MaxFileLength)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/arguments-parameters.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestReadPipelineFile_MultifileTarball(t *testing.T) {
	file, _ := os.Open("test/pipeline_plus_component/pipeline_plus_component.tar.gz")
	pipelineFile, err := ReadPipelineFile("pipeline_plus_component.ai-hub-package", file, MaxFileLength)
	assert.Nil(t, err)

	expectedPipelineFile, _ := ioutil.ReadFile("test/pipeline_plus_component/pipeline.yaml")
	assert.Equal(t, expectedPipelineFile, pipelineFile)
}

func TestReadPipelineFile_UnknownFileFormat(t *testing.T) {
	file, _ := os.Open("test/unknown_format.foo")
	_, err := ReadPipelineFile("unknown_format.foo", file, MaxFileLength)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unexpected pipeline file format")
}

func TestValidateExperimentResourceReference(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	assert.Nil(t, ValidateExperimentResourceReference(manager, validReference))
}

func TestValidateExperimentResourceReference_MoreThanOneRef(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	references := []*apiv1beta1.ResourceReference{
		{
			Key: &apiv1beta1.ResourceKey{
				Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "123"},
			Relationship: apiv1beta1.Relationship_OWNER,
		},
		{
			Key: &apiv1beta1.ResourceKey{
				Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "456"},
			Relationship: apiv1beta1.Relationship_OWNER,
		},
	}
	err := ValidateExperimentResourceReference(manager, references)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "more resource references than expected")
}

func TestValidateExperimentResourceReference_UnexpectedType(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	references := []*apiv1beta1.ResourceReference{
		{
			Key: &apiv1beta1.ResourceKey{
				Type: apiv1beta1.ResourceType_UNKNOWN_RESOURCE_TYPE, Id: "123"},
			Relationship: apiv1beta1.Relationship_OWNER,
		},
	}
	err := ValidateExperimentResourceReference(manager, references)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unexpected resource type")
}

func TestValidateExperimentResourceReference_EmptyID(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	references := []*apiv1beta1.ResourceReference{
		{
			Key: &apiv1beta1.ResourceKey{
				Type: apiv1beta1.ResourceType_EXPERIMENT},
			Relationship: apiv1beta1.Relationship_OWNER,
		},
	}
	err := ValidateExperimentResourceReference(manager, references)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Resource ID is empty")
}

func TestValidateExperimentResourceReference_UnexpectedRelationship(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	references := []*apiv1beta1.ResourceReference{
		{
			Key: &apiv1beta1.ResourceKey{
				Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "123"},
			Relationship: apiv1beta1.Relationship_CREATOR,
		},
	}
	err := ValidateExperimentResourceReference(manager, references)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unexpected relationship for the experiment")
}

func TestValidateExperimentResourceReference_ExperimentNotExist(t *testing.T) {
	clients := resource.NewFakeClientManagerOrFatal(util.NewFakeTimeForEpoch())
	manager := resource.NewResourceManager(clients)
	defer clients.Close()
	err := ValidateExperimentResourceReference(manager, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to get experiment")
}

func TestValidatePipelineSpecAndResourceReferences_WorkflowManifestAndPipelineVersion(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{
		WorkflowManifest: testWorkflow.ToStringForStore()}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReferencesOfExperimentAndPipelineVersion)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Please don't specify a pipeline version or pipeline ID when you specify a workflow manifest or pipeline manifest.")
}

func TestValidatePipelineSpecAndResourceReferences_WorkflowManifestAndPipelineID(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{
		PipelineId:       resource.DefaultFakeUUID,
		WorkflowManifest: testWorkflow.ToStringForStore()}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Please don't specify a pipeline version or pipeline ID when you specify a workflow manifest or pipeline manifest.")
}

func TestValidatePipelineSpecAndResourceReferences_InvalidWorkflowManifest(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{WorkflowManifest: "I am an invalid manifest"}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid argo workflow format.")
}

func TestValidatePipelineSpecAndResourceReferences_NilPipelineSpecAndEmptyPipelineVersion(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	err := ValidatePipelineSpecAndResourceReferences(manager, nil, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Please specify a pipeline by providing a (workflow manifest or pipeline manifest) or (pipeline id or/and pipeline version).")
}

func TestValidatePipelineSpecAndResourceReferences_EmptyPipelineSpecAndEmptyPipelineVersion(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Please specify a pipeline by providing a (workflow manifest or pipeline manifest) or (pipeline id or/and pipeline version).")
}

func TestValidatePipelineSpecAndResourceReferences_InvalidPipelineId(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{PipelineId: "not-found"}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Get pipelineId failed.")
}

func TestValidatePipelineSpecAndResourceReferences_InvalidPipelineVersionId(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	err := ValidatePipelineSpecAndResourceReferences(manager, nil, referencesOfInvalidPipelineVersion)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Get pipelineVersionId failed.")
}

func TestValidatePipelineSpecAndResourceReferences_PipelineIdNotParentOfPipelineVersionId(t *testing.T) {
	clients := initWithExperimentsAndTwoPipelineVersions(t)
	manager := resource.NewResourceManager(clients)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{
		PipelineId: resource.NonDefaultFakeUUID}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReferencesOfExperimentAndPipelineVersion)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "pipeline ID should be parent of pipeline version.")
}

func TestValidatePipelineSpecAndResourceReferences_ParameterTooLongWithPipelineId(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	var params []*apiv1beta1.Parameter
	// Create a long enough parameter string so it exceed the length limit of parameter.
	for i := 0; i < 10000; i++ {
		params = append(params, &apiv1beta1.Parameter{Name: "param2", Value: "world"})
	}
	spec := &apiv1beta1.PipelineSpec{PipelineId: resource.DefaultFakeUUID, Parameters: params}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The input parameter length exceed maximum size")
}

func TestValidatePipelineSpecAndResourceReferences_ParameterTooLongWithWorkflowManifest(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	var params []*apiv1beta1.Parameter
	// Create a long enough parameter string so it exceed the length limit of parameter.
	for i := 0; i < 10000; i++ {
		params = append(params, &apiv1beta1.Parameter{Name: "param2", Value: "world"})
	}
	spec := &apiv1beta1.PipelineSpec{WorkflowManifest: testWorkflow.ToStringForStore(), Parameters: params}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The input parameter length exceed maximum size")
}

func TestValidatePipelineSpecAndResourceReferences_ValidPipelineIdAndPipelineVersionId(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{
		PipelineId: resource.DefaultFakeUUID}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReferencesOfExperimentAndPipelineVersion)
	assert.Nil(t, err)
}

func TestValidatePipelineSpecAndResourceReferences_ValidWorkflowManifest(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	spec := &apiv1beta1.PipelineSpec{WorkflowManifest: testWorkflow.ToStringForStore()}
	err := ValidatePipelineSpecAndResourceReferences(manager, spec, validReference)
	assert.Nil(t, err)
}

func TestValidatePipelineSource_PipelineId(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()

	recurringRun := &apiv2beta1.RecurringRun{
		PipelineSource: &apiv2beta1.RecurringRun_PipelineId{PipelineId: resource.DefaultFakeUUID},
	}
	err := ValidatePipelineSource(manager, recurringRun)
	assert.Nil(t, err)
}

func TestValidatePipelineSource_PipelineSpec(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()

	pipelineSpec := new(structpb.Struct)
	pipelineSpec.UnmarshalJSON([]byte("PipelineSpec"))
	recurringRun := &apiv2beta1.RecurringRun{
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: pipelineSpec},
	}
	err := ValidatePipelineSource(manager, recurringRun)
	assert.Nil(t, err)
}

func TestValidatePipelineSource_EmptyPipelineSource(t *testing.T) {
	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
	defer clients.Close()
	recurringRun := &apiv2beta1.RecurringRun{}
	err := ValidatePipelineSource(manager, recurringRun)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "both pipelineId and pipelineSpec are empty")
}

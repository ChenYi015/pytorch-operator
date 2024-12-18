// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pytorch

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kubeflow/tf-operator/pkg/common/jobcontroller"
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/kubeflow/pytorch-operator/pkg/apis/pytorch/v1"
)

const (
	EnvMasterAddr = "MASTER_ADDR"
	EnvMasterPort = "MASTER_PORT"
	EnvWorldSize = "WORLD_SIZE"
	EnvRank = "RANK"
	EnvPythonUnbuffered = "PYTHONUNBUFFERED"

	// EnvPetMasterAddr is the environment variable name for the master address.
	EnvPetMasterAddr = "PET_MASTER_ADDR"
	// EnvPetMasterPort is the environment variable name for the master port.
	EnvPetMasterPort = "PET_MASTER_PORT"
	// EnvPetNprocPerNode is the environment variable name for the number of processes per node.
	EnvPetNprocPerNode = "PET_NPROC_PER_NODE"
	// EnvPetNnodes is the environment variable name for the number of nodes.
	EnvPetNnodes = "PET_NNODES"
	// EnvPetNodeRank is the environment variable name for the rank of nodes.
	EnvPetNodeRank = "PET_NODE_RANK"
)

// setPodTemplateEnv sets the environment variables for the podTemplateSpec.
func setPodTemplateEnv(podTemplateSpec *corev1.PodTemplateSpec, job *v1.PyTorchJob, rtype v1.PyTorchReplicaType, index string) error {
	rank, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	masterAddr := jobcontroller.GenGeneralName(job.Name, strings.ToLower(string(v1.PyTorchReplicaTypeMaster)), strconv.Itoa(0))
	if rtype == v1.PyTorchReplicaTypeMaster {
		if rank != 0 {
			return errors.New("invalid config: There should be only a single master with index=0")
		}
	} else {
		rank = rank + 1
	}

	masterPort, err := GetPortFromPyTorchJob(job, v1.PyTorchReplicaTypeMaster)
	if err != nil {
		return err
	}

	for i := range podTemplateSpec.Spec.Containers {
		masterAddrFound := false
		masterPortFound := false
		worldSizeFound := false
		pythonUnbufferedFound := false
		petMasterAddrFound := false
		petMasterPortFound := false
		petNprocPerNodeFound := false
		petNnodesFound := false
		petNodeRankFound := false

		for _, env := range podTemplateSpec.Spec.Containers[i].Env {
			switch env.Name {
			case EnvMasterAddr:
				masterAddrFound = true
			case EnvMasterPort:
				masterPortFound = true
			case EnvWorldSize:
				worldSizeFound = true
			case EnvPythonUnbuffered:
				pythonUnbufferedFound = true
			case EnvPetMasterAddr:
				petMasterAddrFound = true
			case EnvPetMasterPort:
				petMasterPortFound = true
			case EnvPetNprocPerNode:
				petNprocPerNodeFound = true
			case EnvPetNnodes:
				petNnodesFound = true
			case EnvPetNodeRank:
				petNodeRankFound = true
			}
		}

		if !masterAddrFound {
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvMasterAddr,
				Value: masterAddr,
			})
		}

		if !masterPortFound {
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name: EnvMasterPort,
				Value: strconv.Itoa(int(masterPort)),
			})
		}

		if !worldSizeFound {
			worldSize := getWorldSize(job)
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name: EnvWorldSize,
				Value: strconv.Itoa(worldSize),
			})
		}

		podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
			Name: EnvRank,
			Value: strconv.Itoa(rank),
		})

		if !pythonUnbufferedFound {
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvPythonUnbuffered,
				Value: "1",
			})
		}

		if !petMasterAddrFound {
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvPetMasterAddr,
				Value: masterAddr,
			})
		}

		if !petMasterPortFound {
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvPetMasterPort,
				Value: strconv.Itoa(int(masterPort)),
			})
		}

		if !petNprocPerNodeFound && job.Spec.NprocPerNode != nil {
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvPetNprocPerNode,
				Value: *job.Spec.NprocPerNode,
			})
		}

		if !petNnodesFound {
			nNodes := int(getTotalReplicas(job))
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvPetNnodes,
				Value: strconv.Itoa(nNodes),
			})
		}

		if !petNodeRankFound {
			nodeRank := rank
			podTemplateSpec.Spec.Containers[i].Env = append(podTemplateSpec.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  EnvPetNodeRank,
				Value: strconv.Itoa(nodeRank),
			})
		}
	}

	return nil
}
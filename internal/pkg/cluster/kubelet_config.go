/*
Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>

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

package cluster

import (
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

// KubeletConfig returns a default kubelet config
func (cluster *Cluster) KubeletConfig() (string, error) {
	kubeletConfig := kubeletconfigv1beta1.KubeletConfiguration{}
	return marshalKubeletConfig(&kubeletConfig)
}

func marshalKubeletConfig(kubeletConfig *kubeletconfigv1beta1.KubeletConfiguration) (string, error) {
	scheme := runtime.NewScheme()
	if err := kubeletconfigv1beta1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, kubeletconfigv1beta1.SchemeGroupVersion)
	encodedKubeletConfig, err := runtime.Encode(encoder, kubeletConfig)
	if err != nil {
		return "", errors.Wrap(err, "could not create a kubelet config")
	}
	return string(encodedKubeletConfig), nil
}

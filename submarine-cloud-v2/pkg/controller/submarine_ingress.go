/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"
	"fmt"

	v1alpha1 "github.com/apache/submarine/submarine-cloud-v2/pkg/apis/submarine/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

func newSubmarineServerIngress(submarine *v1alpha1.Submarine, namespace string) *extensionsv1beta1.Ingress {
	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serverName + "-ingress",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(submarine, v1alpha1.SchemeGroupVersion.WithKind("Submarine")),
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: serverName,
										ServicePort: intstr.FromInt(8080),
									},
									Path: "/",
								},
							},
						},
					},
				},
			},
		},
	}
}

// createIngress is a function to create Ingress.
// Reference: https://github.com/apache/submarine/blob/master/helm-charts/submarine/templates/submarine-ingress.yaml
func (c *Controller) createIngress(submarine *v1alpha1.Submarine, namespace string) error {
	klog.Info("[createIngress]")
	serverName := "submarine-server"

	// Step1: Create ServiceAccount
	ingress, ingress_err := c.ingressLister.Ingresses(namespace).Get(serverName + "-ingress")
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(ingress_err) {
		ingress, ingress_err = c.kubeclientset.ExtensionsV1beta1().Ingresses(namespace).Create(context.TODO(),
			newSubmarineServerIngress(submarine, namespace),
			metav1.CreateOptions{})
		klog.Info("	Create Ingress: ", ingress.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if ingress_err != nil {
		return ingress_err
	}

	if !metav1.IsControlledBy(ingress, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, ingress.Name)
		c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	return nil
}

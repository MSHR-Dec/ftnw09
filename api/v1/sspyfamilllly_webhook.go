/*
Copyright 2023.

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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var sspyfamillllylog = logf.Log.WithName("sspyfamilllly-resource")

func (r *SSpyFamilllly) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-mshr-whisffee-net-v1-sspyfamilllly,mutating=true,failurePolicy=fail,sideEffects=None,groups=mshr.whisffee.net,resources=sspyfamillllies,verbs=create;update,versions=v1,name=msspyfamilllly.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &SSpyFamilllly{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SSpyFamilllly) Default() {
	sspyfamillllylog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-mshr-whisffee-net-v1-sspyfamilllly,mutating=false,failurePolicy=fail,sideEffects=None,groups=mshr.whisffee.net,resources=sspyfamillllies,verbs=create;update,versions=v1,name=vsspyfamilllly.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SSpyFamilllly{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SSpyFamilllly) ValidateCreate() (admission.Warnings, error) {
	sspyfamillllylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SSpyFamilllly) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	sspyfamillllylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SSpyFamilllly) ValidateDelete() (admission.Warnings, error) {
	sspyfamillllylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

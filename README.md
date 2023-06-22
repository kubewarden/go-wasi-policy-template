> **WARNING:** this is not the recommended way to write Kubewarden
> policies using Go. Please read [this](https://docs.kubewarden.io/writing-policies/wasi)
> section of the Kubewarden documentation for more information.

This is the template of a plain WASI policy written using Go. The policy is
then compiled with the official Go compiler.

## Known limitations

Technical limitations caused by Go compiler not having a mature
[WASI](https://wasi.dev/) support:

* The policy requires Go 1.21 or later. Currently this is not yet published,
  hence a Go compiler built from the [`master`](https://github.com/golang/go)
  is required
* The size of the policy is bigger than the ones produced by TinyGo
* This policy requires Kubewarden to support the new `wasi` execution mode. This
  mode provides slower evaluation time compared to the traditional `wapc` one.
  Once [this](https://github.com/golang/go/issues/42372) Go issue is addressed, the
  policy will be rewritten to make use of the traditional Kubewarden policy
  interface.

## Usage

This policy can inspect any kind of Kubernetes resource and ensure:

* A list of user defined annotations are not being used by the resource
* A dictionary of user defined annotations are always present

The policy configuration has the following entries:
* `requiredAnnotations`: a dictionary with a list of annotations that must
  be defined inside of the resource. If not defined, these annotations will
  be added by the policy
* `forbiddenAnnotations`: list of annotations that are not allowed. The
  admission request will be rejected if the resource has any of these annotations

### Example

Given the following configuration:

```yaml
requiredAnnotations:
  cc-center: marketing
  priority: low
forbiddenAnnotations:
- team
- squad
```

All the Kubernetes resources will have the following annotations:

* `cc-center`, with value `marketing`
* `priority`, with value `low`

It's also not going to be allowed to create resources that have either
the `team` or the `squad` annotations set.

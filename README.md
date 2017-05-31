# VMRegistry

**This is not an official Google product**

VMRegistry is a simple GRPC-based API around libvirt that allows to query VM
details and to manage VM state.

You can find all the currently exposed APIs in `proto/vmregistry.proto`.

## Accessing VMRegistry

VMRegistry auth is based on JWT as provided by [credstore](fixme). Consult
credstore documentation on how to generate a token.

There's no RBAC at the moment, so anyone holding a valid token has full access
to the vmregistry, possibly meaning a transitive root access to the host node
via libvirt.

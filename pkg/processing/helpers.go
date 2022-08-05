package processing

const Namespace = "namespace"
const Job = "job"

func generateIdentifier(namespace, name string) string {
	return namespace + "/" + name
}

package ptree

type Scanner interface {
	Scan(UID, QOS string) (Pod, error)
}

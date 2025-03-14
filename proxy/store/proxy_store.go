package store

type ProxyStore interface {
	Get(iou string) (string, bool)
	Set(iou, pgt string) error
	Delete(iou string) error
	Clear() error
}

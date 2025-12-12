package prosemirror

type NodeService struct{}

func NewNodeService() (*NodeService, error) {
	return &NodeService{}, nil
}

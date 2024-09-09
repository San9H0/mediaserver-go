package servers

import (
	"mediaserver-go/dto"
	"mediaserver-go/hubs"
)

type FileServer struct {
	hubManager *hubs.Manager
}

func NewFileServer(hubManager *hubs.Manager) (FileServer, error) {
	return FileServer{
		hubManager: hubManager,
	}, nil
}

func (w *FileServer) StartSession(req dto.EgressFileRequest) (dto.EgressFileResponse, error) {
	//ioBuffer := buffers.NewMemoryBuffer()
	//fileSession, err := sessions.NewFileSession(req.Path, nil, 0, ioBuffer)
	//if err != nil {
	//	return dto.EgressFileResponse{}, err
	//}
	//ctx, cancel := context.WithCancel(context.Background())
	//_ = cancel
	//
	//go fileSession.Run(ctx, nil)
	return dto.EgressFileResponse{}, nil
}

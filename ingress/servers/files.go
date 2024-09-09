package servers

import (
	"context"
	"mediaserver-go/dto"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
)

type FileServer struct {
	hubManager *hubs.Manager
}

func NewFileServer(hubManager *hubs.Manager) (FileServer, error) {
	return FileServer{
		hubManager: hubManager,
	}, nil
}

func (f *FileServer) StartSession(req dto.IngressFileRequest) (dto.IngressFileResponse, error) {
	fileSession, err := sessions.NewFileSession(req.Path, req.MediaTypes, req.Live)
	if err != nil {
		return dto.IngressFileResponse{}, err
	}

	//hub, err := f.hubManager.NewHub("test", types.MediaTypeFromPion(onTrack.remote.Kind()))
	//if err != nil {
	//	fmt.Println("err:", err)
	//	continue
	//}

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	go fileSession.Run(ctx)
	return dto.IngressFileResponse{}, nil
}

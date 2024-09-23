package hubs

//
//const (
//	bufferSize = 32768
//)
//
//func NewWriter(container containers.Container) (*Writer, error) {
//	writer := &Writer{
//		container:  container,
//		tempBuffer: make([]byte, bufferSize),
//		ioBuffer:   buffers.NewMemoryBuffer(),
//
//		pkt: avcodec.AvPacketAlloc(),
//	}
//	if err := writer.init(); err != nil {
//		return nil, err
//	}
//	return writer, nil
//}
//
//type Writer struct {
//	container       containers.Container
//	outputFormatCtx *avformat.FormatContext
//
//	ioBuffer   io.ReadWriteSeeker
//	tempBuffer []byte
//	pkt        *avcodec.Packet
//}
//
//func (w *Writer) init() error {
//	outputFormatCtx := avformat.NewAvFormatContextNull()
//	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", fmt.Sprintf("output.%s", w.container.Extension())); ret < 0 {
//		return errors.New("avformat context allocation failed")
//	}
//	w.outputFormatCtx = outputFormatCtx
//
//	for index, codec := range w.container.Codecs() {
//		outputStream := outputFormatCtx.AvformatNewStream(nil)
//		if outputStream == nil {
//			return errors.New("avformat stream allocation failed")
//		}
//		avCodec := avcodec.AvcodecFindEncoder(types.CodecIDFromType(codec.CodecType()))
//		if avCodec == nil {
//			return errors.New("encoder not found")
//		}
//		avCodecCtx := avCodec.AvCodecAllocContext3()
//		if avCodecCtx == nil {
//			return errors.New("codec context allocation failed")
//		}
//		codec.SetCodecContext(avCodecCtx)
//		if ret := avCodecCtx.AvCodecOpen2(avCodec, nil); ret < 0 {
//			return errors.New("codec open failed")
//		}
//		if ret := avcodec.AvCodecParametersFromContext(outputStream.CodecParameters(), avCodecCtx); ret < 0 {
//			return errors.New("codec parameters from context failed")
//		}
//	}
//
//	avioCtx := avformat.AVIoAllocContext(w.outputFormatCtx, w.ioBuffer, &w.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
//	w.outputFormatCtx.SetPb(avioCtx)
//
//	if err := w.container.SetWriteHeader(w.outputFormatCtx); err != nil {
//		return fmt.Errorf("set write header failed: %w", err)
//	}
//	return nil
//}
//
//func (w *Writer) WritePacket(unit units.Unit) {
//
//}
//
//func (f *HLSSession) readTrack(ctx context.Context, index int, track *hubs.HubSource) error {
//	consumerCh := track.AddConsumer()
//	defer func() {
//		track.RemoveConsumer(consumerCh)
//	}()
//	outputStream := f.outputFormatCtx.Streams()[index]
//	writer := files.NewAVPacketWriter(index, outputStream.TimeBase().Den(), track.MediaType(), track.CodecType())
//
//	for {
//		select {
//		case <-ctx.Done():
//			return nil
//		case unit, ok := <-consumerCh:
//			if !ok {
//				return nil
//			}
//
//			if v.basePTS == 0 {
//				v.basePTS = unit.PTS
//			}
//			pts := unit.PTS - v.basePTS
//			if v.baseDTS == 0 {
//				v.baseDTS = unit.DTS
//			}
//			dts := unit.DTS - v.baseDTS
//
//			if v.filter.Drop(unit) {
//				return nil
//			}
//
//			flag := 0
//			if v.filter.KeyFrame(unit) {
//				flag = 1
//			}
//
//			inputTimebase := avutil.NewRational(1, unit.TimeBase)
//			outputTimebase := avutil.NewRational(1, v.timebase)
//			pkt.SetPTS(avutil.AvRescaleQRound(pts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
//			pkt.SetDTS(avutil.AvRescaleQRound(dts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
//			pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, inputTimebase, outputTimebase))
//			pkt.SetStreamIndex(v.index)
//
//			data := v.bitStream.SetBitStream(unit)
//
//			pkt.SetData(data)
//			pkt.SetFlag(flag)
//			return pkt
//
//			f.mu.Lock()
//			_ = f.outputFormatCtx.AvInterleavedWriteFrame(pkt)
//			f.mu.Unlock()
//			pkt.AvPacketUnref()
//		}
//	}
//}

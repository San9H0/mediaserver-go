#include <stdio.h>
#include <libavformat/avformat.h>

void hello() {
    AVFormatContext* fmt_ctx = NULL;
    printf("Hello from C in another file");
}
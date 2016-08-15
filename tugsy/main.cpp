#include <sdl.h>
#include <thread>


int main() {
    int ret = init(sdl_state);
    shutdown(sdl_state);

    return ret;
}

#include <Tugsy.h>

Tugsy::Tugsy(SdlState &state) {
    this->sdlState = state;
}

int Tugsy::onExecute(){

    // We never explicitly exit, instead relying on our ability to
    // handle a signal appropriately
    while(running) {
        SDL_Event event;
        while(SDL_PollEvent(&event)) {
            onEvent(&event);
        }

        onLoop();
        onRender();
    }

    onCleanup();

    return 0;
}

void Tugsy::onEvent(SDL_Event* event) {
}

void Tugsy::onLoop() {
}

void Tugsy::onRender() {
}

void Tugsy::onCleanup() {
    // Nothing to be done... the SDL destructor will take care of
    // the interface and we have no state to manage
}

void sigHandler(int param) {
    running = false;
}

int main(int argc, char* argv[]) {
    if (signal (SIGINT, sigHandler) > 0) {
        cerr << "Failed to load the SIGINT handler" << std::endl;
        return 1;
    }

    SdlState sdlState = SdlState();
    Tugsy app = Tugsy(sdlState);
    sdlState.loadView();
    return app.onExecute();
}

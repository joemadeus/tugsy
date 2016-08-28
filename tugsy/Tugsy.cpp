#include <Tugsy.h>

Tugsy::Tugsy(SdlState &state) {
    this->sdlState = state;
}

int Tugsy::OnExecute(){

    // We never explicitly exit, instead relying on our ability to
    // handle a signal or cntl-c appropriately
    while(true) {
        SDL_Event event;
        while(SDL_PollEvent(&event)) {
            OnEvent(&event);
        }

        OnLoop();
        OnRender();
    }

    OnCleanup();

    return 0;
}

void Tugsy::OnEvent(SDL_Event* event) {
}

void Tugsy::OnLoop() {
}

void Tugsy::OnRender() {
}

void Tugsy::OnCleanup() {
    // Nothing to be done... the SDL destructor will take care of
    // the interface and we have no state to manage
}

int main(int argc, char* argv[]) {
    SdlState sdlState = SdlState();
    sdlState.loadView();

    Tugsy app = Tugsy(sdlState);
    return app.OnExecute();
}

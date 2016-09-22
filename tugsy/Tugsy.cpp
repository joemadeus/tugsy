#include "Tugsy.h"

Tugsy::Tugsy(SdlState &state, PositionData &positions) {
    this->sdlState = state;
    this->positionData = positions;

    for (std::string v = KNOWN_VIEWS.cbegin(); v != KNOWN_VIEWS.cend(); v++) {
        this->views[i] = View(*v, state);
    }

    this->currentViewIndex = 0;
}

int Tugsy::onExecute() {
    while(running) {
        SDL_Event* event;
        while(SDL_PollEvent(event)) {
            onEvent(event);
        }

        onLoop();
        onRender();
    }

    onCleanup();

    return 0;
}

void Tugsy::onEvent(SDL_Event* event) {
    // TODO: On touch, show the next view
    this->currentViewIndex =
        (this->currentViewIndex >= sizeof(views) - 1)
        ? 0
        : this->currentViewIndex++;

    this->currentView());
}

void Tugsy::onLoop() {
    // TODO: Read from the radio and other sources and write the data
    // to the PositionData instance
}

void Tugsy::onRender() {
    this->currentView().renderPositions(this->positionData.getLatestPositions());
    this->sdlState.flip();
}

void Tugsy::onCleanup() {
    // Nothing to be done... the SDL & View destructors will take care of
    // the interface and we have no other state to manage
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

    return app.onExecute();
}

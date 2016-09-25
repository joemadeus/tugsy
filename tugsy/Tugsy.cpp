#include "Tugsy.h"

Tugsy::Tugsy(SdlState &state, PositionData &positions) {
    this->sdlState = state;
    this->positionData = positions;

    unsigned int i = 0;
    for (std::string viewName : KNOWN_VIEWS) {
        this->views[i++] = View(viewName, state);
    }
}

int Tugsy::onExecute() {
    while(running) {
        SDL_Event* event = NULL;
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
    switch (event->type) {
        case SDL_KEYDOWN:
            // Any keypress will switch to the next view
            // in the cycle
            this->currentViewIndex =
                (this->currentViewIndex >= this->views.size() - 1)
                ? 0
                : this->currentViewIndex + 1;

            this->currentView().rebuild();
            break;

        case SDL_FINGERDOWN:
            // TODO: On touch to the touchscreen, show vessel info
            // see SDL_TouchFingerEvent
            // vessel_info_t vessel =
            //     this->currentView().whatsHere(
            //         (int) round(x * SCREEN_WIDTH),
            //         (int) round(y * SCREEN_HEIGHT));
            // if (vessel != NULL) {
            // }

            break;

        default:
            // do nothing
            break;
    }
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

View& Tugsy::currentView() {
    return this->views[this->currentViewIndex];
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
    PositionData positionData = PositionData();
    Tugsy app = Tugsy(sdlState, positionData);

    return app.onExecute();
}

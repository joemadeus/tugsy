#include "Tugsy.h"

Tugsy::Tugsy(SdlState &state, PositionData &positions) {
    this->sdlState = state;
    this->positionData = positions;
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
    // TODO: On touch, show the next view
}

void Tugsy::onLoop() {
    this->positionData.updateVessels();
    this->positionData.expireVessels();
}

void Tugsy::onRender() {
    vessel_info_t expiredVessels = this->positionData.getExpiredPositions(this->lastUpdate);
    vessel_info_t updatedVessels = this->positionData.getUpdatedPositions(this->lastUpdate);
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

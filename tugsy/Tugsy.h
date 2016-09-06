#ifndef TUGSY_H
#define TUGSY_H

#include <stdio.h>
#include <thread>
#include <signal.h>
#include <SDL2/SDL.h>

#include "sdl.h"
#include "PositionData.h"

using namespace std;

const std::string KNOWN_VIEWS = {
    "pvd_harbor",
    "pvd_to_bristol",
    "pvd_to_gansett"
};

class Tugsy {
public:
    Tugsy(SdlState &state, PositionData &positions);
    int onExecute();
    bool onInit();
    void onEvent(SDL_Event* Event);
    void onLoop();
    void onRender();
    void onCleanup();

private:
    SdlState sdlState;
    PositionData &positionData
    View views[sizeof(KNOWN_VIEWS)];
    int currentView = 0;
};

bool running = true;

void sigHandler(int param);

#endif

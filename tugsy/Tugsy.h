#ifndef TUGSY_H
#define TUGSY_H

#include <array>
#include <stdio.h>
#include <thread>
#include <signal.h>

#include "sdl.h"
#include "PositionData.h"
#include "View.h"

using namespace std;

class Tugsy {
private:
    Tugsy(SdlState &state, PositionData &positions);
    int onExecute();
    bool onInit();
    void onEvent(SDL_Event* Event);
    void onLoop();
    void onRender();
    void onCleanup();

    View& currentView();

    SdlState sdlState;
    PositionData &positionData;
    std::array<View> views;
    int currentViewIndex = 0;
    long lastUpdateMillis = 0;
};

bool running = true;

void sigHandler(int param);

#endif

#include "View.h"

View::View(std::string &viewName, SdlState &state) {
    this->viewName = viewName;
    this->background = state.initTexture(getResource(viewName, BACKGROUND_FILE_NAME));
    this->marker_sheet = state.initTexture(getResource(viewName, MARKER_SHEET_FILE_NAME));

    // TODO: Load the geo coords of the view's four corners from
    // a property file, then use them to assign the local geo vars
}

View::~View() {
    cleanup(this->background, this->marker_sheet);
}

marker_dims_t View::getMarker(int markerType const, int vesselType const) {
    marker_dims_t dims = NULL;
    dims.x = markerType * MARKER_WIDTH;
    dims.y = vesselType * MARKER_HEIGHT;
    dims.w = dims.x + MARKER_WIDTH;
    dims.h = dims.y + MARKER_HEIGHT;

    return dims;
}

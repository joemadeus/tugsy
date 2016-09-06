#include "sdl.h"

SdlState::SdlState() {
	initContext();
}

SdlState::~SdlState() {
	cleanup(this->window, this->renderer);
	IMG_Quit();
	SDL_Quit();
}

void SdlState::initContext() {
	// SDL_GetBasePath will return NULL if something went wrong in retrieving the path
	char *basePath = SDL_GetBasePath();
	if ( ! basePath) {
		throw exception(sdlError("Error getting resource path: "));

	baseRes = basePath;
	SDL_free(basePath);

	// We replace the last bin/ with res/ to get the resource path
	size_t pos = baseRes.rfind("bin");
	baseRes = baseRes.substr(0, pos) + "res" + PATH_SEP;

	if (SDL_Init(SDL_INIT_VIDEO) != 0) {
		throw exception(sdlError("SDL_Init"));
	}

	if ((IMG_Init(IMG_INIT_PNG) & IMG_INIT_PNG) != IMG_INIT_PNG) {
		std::string const msg = sdlError("IMG_Init")
		SDL_Quit();
		throw exception(msg);
	}

	this->window = SDL_CreateWindow(
		"tugsy-dev", SCREEN_X, SCREEN_Y, SCREEN_WIDTH, SCREEN_HEIGHT, SDL_WINDOW_SHOWN);
	if (this->window == nullptr) {
		std::string const msg = sdlError("CreateWindow")
		IMG_Quit();
		SDL_Quit();
		throw exception(msg);
	}

	this->renderer = SDL_CreateRenderer(
		this->window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
	if (this->renderer == nullptr) {
		std::string const msg = sdlError("CreateRenderer")
		cleanup(this->window);
		IMG_Quit();
		SDL_Quit();
		throw exception(msg);
	}
}

SDL_Texture* SdlState::initTexture(const std::string &filename) {
	SDL_Texture *texture = IMG_LoadTexture(this->renderer, filename.c_str());
	if (texture == nullptr){
		throw exception(sdlError("Couldn't load file " + filename + " to a texture."));
	}
	return texture;
}

bool SdlState::drawNextView() {
	this->currentView =
        (this->currentView >= sizeof(views) - 1)
        ? 0
        : this->currentView++;

	// TODO: Draw to the renderer

	return true;
}

std::string SdlState::getResource(const std::string &viewName, const std::string &resource) {
	return this->baseRes + PATH_SEP + viewName + PATH_SEP + resource;
}

std::string sdlError(const std::string &msg){
	return msg + " error: " + SDL_GetError();
}

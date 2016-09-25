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
		sdlError("Error getting resource path: ");
		throw exception();
	}

	baseRes = basePath;
	SDL_free(basePath);

	// We replace the last bin/ with res/ to get the resource path
	size_t pos = baseRes.rfind("bin");
	baseRes = baseRes.substr(0, pos) + "res" + PATH_SEP;

	if (SDL_Init(SDL_INIT_VIDEO) != 0) {
		sdlError("SDL_Init");
		throw exception();
	}

	if ((IMG_Init(IMG_INIT_PNG) & IMG_INIT_PNG) != IMG_INIT_PNG) {
		sdlError("IMG_Init");
		SDL_Quit();
		throw exception();
	}

	this->window = SDL_CreateWindow(
		"tugsy-dev", SCREEN_X, SCREEN_Y, SCREEN_WIDTH, SCREEN_HEIGHT, SDL_WINDOW_SHOWN);
	if (this->window == nullptr) {
		sdlError("CreateWindow");
		IMG_Quit();
		SDL_Quit();
		throw exception();
	}

	this->renderer = SDL_CreateRenderer(
		this->window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
	if (this->renderer == nullptr) {
		sdlError("CreateRenderer");
		cleanup(this->window);
		IMG_Quit();
		SDL_Quit();
		throw exception();
	}
}

SDL_Texture* SdlState::initTexture(const std::string &viewName, const std::string &resource) {
	std::string filename = this->getResource(viewName, resource);
	SDL_Texture *texture = IMG_LoadTexture(this->renderer, filename.c_str());
	if (texture == nullptr){
		sdlError("Couldn't load file " + filename + " to a texture.");
		throw exception();
	}
	return texture;
}

void SdlState::paint(SDL_Texture* texture,
	                 const SDL_Rect* srcrect,
	                 const SDL_Rect* dstrect) {
	if (SDL_RenderCopy(this->renderer, texture, srcrect, dstrect) != 0) {
		sdlError("Could not paint a texture: ");
		// TODO: Better handling for rendering errors?
	}
}

void SdlState::flip() {
	SDL_RenderPresent(this->renderer);
}

std::string SdlState::getResource(const std::string &viewName, const std::string &resource) {
	return this->baseRes + PATH_SEP + viewName + PATH_SEP + resource;
}

void sdlError(const std::string &msg){
	cerr << msg << " error: " << SDL_GetError() << std::endl;
}

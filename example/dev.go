//+build !prod

package main

func (a *App) Install() {
	a.ConfPath = "./conf"
}

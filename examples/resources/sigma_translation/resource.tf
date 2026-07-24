resource "sigma_translation" "example" {
  lng = "fr"
  translations = {
    Hello = "Bonjour"
    World = "Monde"
  }
}

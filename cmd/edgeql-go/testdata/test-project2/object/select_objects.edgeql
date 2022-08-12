select schema::Function {
  name,
  language,
  params: {
    name,
    default,
  }
}

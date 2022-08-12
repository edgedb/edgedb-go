select schema::Function {
  name,
  language,
  params: {
    name,
    default,
  }
}
limit 1;

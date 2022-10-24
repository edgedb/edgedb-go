create type Person {
  create required property Name -> str {
    create constraint exclusive;
  };
  create multi link Friends -> Person {
    create property Strength -> float64;
  }
};

select Person {
  Name,
  Friends: {
    Name,
    @Strength,
  },
}


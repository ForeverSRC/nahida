rsconf = {
    _id: "mdbDefGuide",
    members: [
      {_id: 0, host: "mongo-rs0:27017"},
      {_id: 1, host: "mongo-rs1:27018"},
      {_id: 2, host: "mongo-rs2:27019"}
    ]
}

rs.initiate(rsconf)

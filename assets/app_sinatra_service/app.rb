require "sinatra"
require "pg"
require "json"

get "/test" do
  services = JSON.parse(ENV["VCAP_SERVICES"])
  credentials = services["a.postgresql"][0]["credentials"]
  begin
    session = PG.connect(
      user: credentials["username"],
      password: credentials["password"],
      host:    credentials["host"],
      port:     credentials["port"],
      dbname: credentials["dbname"]
    )
    session.exec("SELECT CURRENT_TIME")
    session.finish
    "works"
  rescue => e
    e.inspect
  end
end


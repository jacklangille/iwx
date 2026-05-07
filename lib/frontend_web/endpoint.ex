defmodule FrontendWeb.Endpoint do
  use Phoenix.Endpoint, otp_app: :frontend

  socket "/socket", FrontendWeb.UserSocket,
    websocket: true,
    longpoll: false

  plug Plug.Static,
    at: "/",
    from: :frontend,
    gzip: false,
    only: ~w(assets)

  if code_reloading? do
    socket "/phoenix/live_reload/socket", Phoenix.LiveReloader.Socket
    plug Phoenix.CodeReloader
    plug Phoenix.LiveReloader
  end

  plug Plug.Parsers,
    parsers: [:urlencoded, :multipart, :json],
    pass: ["*/*"],
    json_decoder: Jason

  plug Plug.RequestId
  plug Plug.Head
  plug FrontendWeb.Router
end

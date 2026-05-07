defmodule FrontendWeb.StationExplorerController do
  use FrontendWeb, :controller

  def index(conn, _params) do
    render(conn, :index)
  end
end

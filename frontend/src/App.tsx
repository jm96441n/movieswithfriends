import "./App.css";
import { Theme } from "react-daisyui";
import { Outlet } from "@tanstack/react-router";
import Nav from "./Nav.tsx";

function App() {
  return (
    <Theme dataTheme="retro">
      <div className="App">
        <Nav />
        <Outlet />
      </div>
    </Theme>
  );
}

export default App;

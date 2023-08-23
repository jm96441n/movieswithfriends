import React, { useState } from "react";
import "./App.css";
import {
  AuthenticatedContext,
  AuthenticatedDispatchContext,
} from "./Context.tsx";
import { Theme } from "react-daisyui";
import { Outlet } from "@tanstack/react-router";
import Nav from "./Nav.tsx";

function initialAuthValue() {
  const isAuthenticated = localStorage.getItem("moviesauth");
  console.log(isAuthenticated);
  if (isAuthenticated == null || isAuthenticated === "false") {
    return false;
  }
  return true;
}

function App() {
  const [authenticated, setAuthenticated] = useState(initialAuthValue());
  return (
    <Theme dataTheme="synthwave">
      <AuthenticatedContext.Provider value={authenticated}>
        <AuthenticatedDispatchContext.Provider value={setAuthenticated}>
          <div className="App h-screen">
            <Nav authenticated={authenticated} />
            <Outlet />
          </div>
        </AuthenticatedDispatchContext.Provider>
      </AuthenticatedContext.Provider>
    </Theme>
  );
}

export default App;

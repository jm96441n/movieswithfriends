import React, { useState, useContext } from "react";
import { Button } from "react-daisyui";
import { useNavigate } from "@tanstack/react-router";
import { AuthenticatedDispatchContext } from "./Context.tsx";

function Logoutbutton() {
  const [loading, setLoading] = useState(false);
  const [disabled, setDisabled] = useState(false);

  const navigate = useNavigate({ from: "/" });
  const authenticatedDispatch = useContext(AuthenticatedDispatchContext);

  function handleLogout(e: ChangeEvent<HTMLInputElement>) {
    e.preventDefault();
    setLoading(true);
    setDisabled(true);
    const logout = async () => {
      try {
        const response = await fetch("http://localhost:8080/logout", {
          method: "POST",
          mode: "cors",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
        });
        if (response.ok) {
          localStorage.removeItem("moviesauth");
          authenticatedDispatch(false);
          navigate({ to: "/" });
        }
      } catch (error) {
        console.error(error);
      }
    };
    logout();
  }

  return (
    <Button
      color="accent"
      size="sm"
      onClick={handleLogout}
      loading={loading}
      disabled={disabled}
    >
      Logout
    </Button>
  );
}

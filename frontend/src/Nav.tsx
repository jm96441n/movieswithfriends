import React, { useState, useContext } from "react";
import { Navbar, Button, Menu } from "react-daisyui";
import { Link, useNavigate } from "@tanstack/react-router";
import { AuthenticatedDispatchContext } from "./Context.tsx";


function Nav({ authenticated }) {
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
          localStorage.removeItem("moviesauth")
          authenticatedDispatch(false);
          navigate({ to: "/" });
        }
      } catch (error) {
        console.error(error);
      }
    }
    logout();
  }

  function menuItems() {
    if (authenticated) {
      return (
        <>
          <Menu.Item>
            <Link to="/profile">Profile</Link>
          </Menu.Item>
          <Menu.Item>
            <Button
              color="accent"
              size="sm"
              onClick={handleLogout}
              loading={loading}
              disabled={disabled}
            >
              Logout
            </Button>
          </Menu.Item>
        </>
      )
    }

    return (
      <>
        <Menu.Item>
          <Link to="/login">Login</Link>
        </Menu.Item>
        <Menu.Item>
          <Link to="/signup">Signup</Link>
        </Menu.Item>
      </>
    )
  }

  return (
    <Navbar>
      <Menu horizontal="true">
        <Menu.Item>
          <Link to="/">
            <Button className="text-xl normal-case" color="ghost">
              MoviesWithFriends
            </Button>
          </Link>
        </Menu.Item>
        {menuItems()}
      </Menu>
    </Navbar>
  );
}

export default Nav;

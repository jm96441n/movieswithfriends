import React from "react";
import { Navbar, Button, Menu } from "react-daisyui";
import { Link } from "@tanstack/react-router";
import LogoutButton from "./Login.tsx";

function Nav({ authenticated }) {
  function menuItems() {
    if (authenticated) {
      return (
        <>
          <Menu.Item>
            <Link to="/profile">Profile</Link>
          </Menu.Item>
          <Menu.Item>
            <LogoutButton />
          </Menu.Item>
        </>
      );
    }

    return (
      <>
        <Navbar.End>
          <Menu horizontal="true">
            <Menu.Item>
              <Link to="/login">Login</Link>
            </Menu.Item>
            <Menu.Item>
              <Link to="/signup">Signup</Link>
            </Menu.Item>
          </Menu>
        </Navbar.End>
      </>
    );
  }

  return (
    <Navbar>
      <Navbar.Start>
        <Menu horizontal="true">
          <Menu.Item>
            <Link to="/">
              <Button className="text-xl normal-case" color="ghost">
                MoviesWithFriends
              </Button>
            </Link>
          </Menu.Item>
        </Menu>
      </Navbar.Start>
      {menuItems()}
    </Navbar>
  );
}

export default Nav;

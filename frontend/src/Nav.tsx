import React from "react";
import { Navbar, Button, Menu } from "react-daisyui";
import { Link } from "@tanstack/react-router";

function Nav({ authenticated }) {
  function menuItems() {
    if (authenticated) {
      return (
        <>
          <Menu.Item>
            <Link to="/profile">Profile</Link>
          </Menu.Item>
          <Menu.Item>
            <Link to="/logout">Logout</Link>
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

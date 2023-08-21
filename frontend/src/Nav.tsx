import { Navbar, Button, Menu } from "react-daisyui";
import { Link } from "@tanstack/react-router";

function Nav() {
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
        <Menu.Item>
          <Link to="/login">Login</Link>
        </Menu.Item>
        <Menu.Item>
          <Link to="/signup">Signup</Link>
        </Menu.Item>
      </Menu>
    </Navbar>
  );
}

export default Nav;

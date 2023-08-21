import { Router, Route, RootRoute } from "@tanstack/react-router";
import App from "./App.tsx";
import Login from "./Login.tsx";
import Signup from "./Signup.tsx";
import Profile from "./Profile.tsx";

const rootRoute = new RootRoute({ component: App });

const indexRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/",
  component: Index,
});

function Index() {
  return (
    <div>
      <h3>Welcome Home!</h3>
    </div>
  );
}

const loginRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: Login,
});

const signupRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/signup",
  component: Signup,
});

const profileRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/profile",
  component: Profile,
  loader: async () => {
    const res = await fetch("http://localhost:8080/profile", {
      credentials: "include",
      mode: "cors",

      headers: {
        "Content-Type": "application/json",
      },
    });
    console.log(res.status);
    if (!res.ok) throw new Error("Failed to fetch posts");
    const body = await res.json();
    return body;
  },
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  signupRoute,
  profileRoute,
]);

const MyRouter = new Router({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof MyRouter;
  }
}

export default MyRouter;

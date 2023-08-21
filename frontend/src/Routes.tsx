import { Router, Route, RootRoute } from "@tanstack/react-router";
import App from "./App.tsx";
import Login from "./Login.tsx";
import Signup from "./Signup.tsx";

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

const routeTree = rootRoute.addChildren([indexRoute, loginRoute, signupRoute]);

const MyRouter = new Router({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof MyRouter;
  }
}

export default MyRouter;

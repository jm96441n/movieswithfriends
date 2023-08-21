import React, { useState } from 'react';
import { Form, Input, Join, Button } from "react-daisyui";

function Signup() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [partyID, setPartyID] = useState("");

  function handleOnClick() {
    alert(`${email} ${password} ${partyID}`)
  }

  return (
    <div>
      <h3>Signup</h3>
      <Form>
        <Join>
          <div className="columns-2">
            <Form.Label title="Email" />
          </div>
          <div className="columns-1">
            <Input type="text" value={email} placeholder="Email" className="input-bordered" onChange={(e) => setEmail(e.target.value)}/>
          </div>
        </Join>
        <Join>
          <div className="columns-2">
            <Form.Label title="Password" />
          </div>
          <div className="columns-1">
            <Input
              type="password"
              placeholder="Password"
              className="input-bordered"
              onChange={(e) => setPassword(e.target.value)}
              value={password}
            />
          </div>
        </Join>
        <Join>
          <Form.Label title="Party ID" />
          <Input
            type="text"
            placeholder="Party ID"
            className="input-bordered"
            onChange={(e) => setPartyID(e.target.value)}
            value={partyID}
          />
        </Join>
          <div className="columns-2">
            <Button tag="button" onClick={handleOnClick}>Sign Up!</Button>
          </div>
      </Form>
    </div>
  );
}

export default Signup;

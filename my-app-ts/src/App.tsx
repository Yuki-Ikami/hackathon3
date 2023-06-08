import React from "react";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Register from "./Register";
import Login from "./Login";
import Mypage from "./Mypage";
import Channel from "./Channel";


const App: React.FC = () => {
  return (
    <div className="container">
      <BrowserRouter>
        <Routes>
          <Route path="/register" element={<Register />} />
          <Route path="/login" element={<Login />} />
          <Route path="/:email" element={<Mypage />} />
          <Route path="/channel/:channel_id/:email" element={<Channel />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
};

export default App;

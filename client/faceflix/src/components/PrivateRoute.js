import React from "react";
import { Navigate } from 'react-router-dom';
import { getCurrentUser } from "../services/auth.service";


const PrivateRoute = ({ children }) => {
    let isLoggedIn = getCurrentUser() !== null;
  
    if (!isLoggedIn) {
      return <Navigate to="/login" replace />;
    }
  
    return children;
  };

export default PrivateRoute;
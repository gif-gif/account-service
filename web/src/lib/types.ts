export type AdminUser = {
  id: string;
  username: string;
};

export type APIErrorBody = {
  error?: {
    code?: string;
    message?: string;
    request_id?: string;
  };
};

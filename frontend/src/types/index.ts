export interface Customer {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  phone_verified: boolean;
  email_verified: boolean;
  status: 'active' | 'suspended' | 'closed';
  preferred_language: string;
  timezone: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface Account {
  id: string;
  customer_id?: string;
  account_number: string;
  account_type: 'checking' | 'savings' | 'loan';
  currency: string;
  status: 'active' | 'frozen' | 'closed';
  created_at: string;
  updated_at: string;
}

export interface AccountBalance {
  account_id: string;
  balance: string;
  currency: string;
  as_of: string;
}

export interface Transaction {
  id: string;
  type: string;
  status: 'pending' | 'completed' | 'failed';
  amount: string;
  currency: string;
  reference?: string;
  created_at: string;
  completed_at?: string;
}

export interface LoginResponse {
  access_token: string;
  expires_at: string;
  token_type: string;
}

export interface ApiError {
  error: string;
}

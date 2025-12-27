import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuth } from '@/context/AuthContext';
import { getAccounts, getAccountBalance, createAccount } from '@/api/client';
import type { Account, AccountBalance } from '@/types';
import { LogOut, Plus, ArrowRightLeft, CreditCard } from 'lucide-react';

interface AccountWithBalance extends Account {
  balance?: AccountBalance;
}

export function DashboardPage() {
  const { logout } = useAuth();
  const [accounts, setAccounts] = useState<AccountWithBalance[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const fetchAccounts = async () => {
    try {
      const accountList = await getAccounts();

      // Fetch balances for each account
      const accountsWithBalances = await Promise.all(
        accountList.map(async (account) => {
          try {
            const balance = await getAccountBalance(account.id);
            return { ...account, balance };
          } catch {
            return account;
          }
        })
      );

      setAccounts(accountsWithBalances);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch accounts');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchAccounts();
  }, []);

  const handleCreateAccount = async () => {
    setIsCreating(true);
    try {
      await createAccount('checking', 'NOK');
      await fetchAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create account');
    } finally {
      setIsCreating(false);
    }
  };

  const handleLogout = async () => {
    await logout();
  };

  const formatCurrency = (amount: string, currency: string) => {
    const num = parseFloat(amount);
    return new Intl.NumberFormat('nb-NO', {
      style: 'currency',
      currency: currency,
    }).format(num);
  };

  const getAccountTypeIcon = (type: string) => {
    switch (type) {
      case 'checking':
        return <CreditCard className="h-5 w-5" />;
      case 'savings':
        return <CreditCard className="h-5 w-5" />;
      default:
        return <CreditCard className="h-5 w-5" />;
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold">Fjord Bank</h1>
          <Button variant="ghost" size="sm" onClick={handleLogout}>
            <LogOut className="h-4 w-4 mr-2" />
            Sign Out
          </Button>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {error && (
          <div className="mb-6 p-4 text-sm text-destructive bg-destructive/10 rounded-md">
            {error}
          </div>
        )}

        {/* Quick Actions */}
        <div className="flex gap-4 mb-8">
          <Button onClick={handleCreateAccount} disabled={isCreating}>
            <Plus className="h-4 w-4 mr-2" />
            {isCreating ? 'Creating...' : 'New Account'}
          </Button>
          <Button variant="outline" asChild>
            <Link to="/transfer">
              <ArrowRightLeft className="h-4 w-4 mr-2" />
              Transfer
            </Link>
          </Button>
        </div>

        {/* Accounts List */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Your Accounts</h2>

          {accounts.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                <CreditCard className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>You don't have any accounts yet.</p>
                <p className="text-sm">Click "New Account" to create your first account.</p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {accounts.map((account) => (
                <Card key={account.id} className="hover:shadow-md transition-shadow">
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {getAccountTypeIcon(account.account_type)}
                        <CardTitle className="text-base capitalize">
                          {account.account_type} Account
                        </CardTitle>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        account.status === 'active'
                          ? 'bg-green-100 text-green-700'
                          : 'bg-gray-100 text-gray-700'
                      }`}>
                        {account.status}
                      </span>
                    </div>
                    <CardDescription className="font-mono text-xs">
                      {account.account_number}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    {account.balance ? (
                      <div className="space-y-1">
                        <p className="text-2xl font-bold">
                          {formatCurrency(account.balance.balance, account.balance.currency)}
                        </p>
                      </div>
                    ) : (
                      <p className="text-muted-foreground">Balance unavailable</p>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

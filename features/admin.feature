Feature: Administration
    Background:
        Given there is an admin account bootstrapped

    Scenario: I use bootstrapped admin credentials to authenticate
        Given bootstrap account credentials are valid
        When I use bootstrap credentials to authenticate
        Then I am successfully authenticated as admin

    Rule: Admin can manage other accounts

        Scenario Outline: I create another account as an admin
            Given I am authenticated as admin
            When I create a new <type> account
            Then account exists

            Examples:
                |type   |
                |user   |
                |admin  |

        Scenario Outline: I delete another account as an admin
            Given I am authenticated as admin
            And there exists another <type> account
            When I delete that account
            Then account no longer exists

            Examples:
                |type   |
                |user   |
                |admin  |

        Scenario: I cannot delete my own account as an admin
            Given I am authenticated as admin
            When I try to delete my account
            Then I get an error
            And my account exists


        Scenario: I cannot delete bootstrapped admin account as an admin
            Given I am authenticated as admin
            When I try to delete bootstrapped admin account
            Then I get an error
            And bootstrapped admin account exists

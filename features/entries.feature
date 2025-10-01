Feature: Entry management
    Background:
        Given there exists a user account
        And there exists a client

    Rule: Users can add entries
        Scenario Outline: I add an entry
            Given I am authenticated as user
            And entry url points to <page> html page
            When I try to add an entry
            Then entry addition <success> succeed
            And the entry <exist> exist

        Examples:
            |page   |success    |exist      |
            |valid  |should     |should     |
            |invalid|should not |should not |

    Rule: Users can view their entries
        Scenario: I view an entry created by me
            Given I am authenticated as user
            And I created a valid entry
            When I view the created entry
            Then entry content should match the content at creation time
